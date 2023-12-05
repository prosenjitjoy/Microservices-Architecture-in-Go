package main

import (
	"context"
	"flag"
	"log"
	"main/discovery"
	"main/discovery/memory"
	"main/rpc"
	"main/utils"
	"net"

	metadatatest "main/metadata/testutil"
	movietest "main/movie/testutil"
	ratingtest "main/rating/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	metadataServiceName = "metadata"
	ratingServiceName   = "rating"
	movieServiceName    = "movie"

	metadataServiceAddr = "localhost:8081"
	ratingServiceAddr   = "localhost:8082"
	movieServiceAddr    = "localhost:8083"
)

func main() {
	var config string
	flag.StringVar(&config, "config", ".env", "Configuration path")
	flag.Parse()
	cfg := utils.LoadConfig(config)

	log.Println("Starting the integration test")

	ctx := context.Background()
	registry := memory.NewRegistry()

	log.Println("Setting up service handlers and clients")

	metadataSrv := startMetadataService(ctx, registry)
	defer metadataSrv.GracefulStop()
	ratingSrv := startRatingService(ctx, registry, cfg)
	defer ratingSrv.GracefulStop()
	movieSrv := startMovieService(ctx, registry)
	defer movieSrv.GracefulStop()

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	metadataConn, err := grpc.Dial(metadataServiceAddr, opts)
	if err != nil {
		panic(err)
	}
	defer metadataConn.Close()
	metadataClient := rpc.NewMetadataServiceClient(metadataConn)

	ratingConn, err := grpc.Dial(ratingServiceAddr, opts)
	if err != nil {
		panic(err)
	}
	defer ratingConn.Close()
	ratingClient := rpc.NewRatingServiceClient(ratingConn)

	movieConn, err := grpc.Dial(movieServiceAddr, opts)
	if err != nil {
		panic(err)
	}
	defer movieConn.Close()
	movieClient := rpc.NewMovieServiceClient(movieConn)

	log.Println("Saving test metadata via metadata service")

	m := &rpc.Metadata{
		MovieId:     "the-movie",
		Title:       "The Movie",
		Description: "The Movie, the one and only",
		Director:    "Mr. D",
	}

	if _, err := metadataClient.PutMetadata(ctx, &rpc.PutMetadataRequest{
		Metadata: m,
	}); err != nil {
		log.Fatalf("put metadata: %v", err)
	}

	log.Println("Retrieving test metadata via metadata service")

	getMetadataResp, err := metadataClient.GetMetadata(ctx, &rpc.GetMetadataRequest{MovieId: m.MovieId})
	if err != nil {
		log.Fatalf("get metadata: %v", err)
	}
	if diff := cmp.Diff(getMetadataResp.Metadata, m, cmpopts.IgnoreUnexported(rpc.Metadata{})); diff != "" {
		log.Fatalf("get metadata after put mismatch: %v", diff)
	}

	log.Println("Getting movie details via movie service")
	wantMovieDetails := &rpc.MovieDetails{
		Metadata: m,
	}

	getMovieDetailsResp, err := movieClient.GetMovieDetails(ctx, &rpc.GetMovieDetailsRequest{MovieId: m.MovieId})
	if err != nil {
		log.Fatalf("geet movie details: %v", err)
	}

	if diff := cmp.Diff(getMovieDetailsResp.MovieDetails, wantMovieDetails, cmpopts.IgnoreUnexported(rpc.MovieDetails{}, rpc.Metadata{})); diff != "" {
		log.Fatalf("get movie details after put mismatch: %v", err)
	}

	log.Println("Saving first rating via rating service")

	const userID = "user0"
	const recordTypeMovie = "movie"
	firstRating := int32(5)
	if _, err := ratingClient.PutRating(ctx, &rpc.PutRatingRequest{
		UserId:      userID,
		RecordId:    m.MovieId,
		RecordType:  recordTypeMovie,
		RatingValue: firstRating,
	}); err != nil {
		log.Fatalf("put rating: %v", err)
	}

	log.Println("Retrieving initial aggregated rating via rating service")

	getAggregatedRatingResp, err := ratingClient.GetAggregatedRating(ctx, &rpc.GetAggregatedRatingRequest{
		RecordId:   m.MovieId,
		RecordType: recordTypeMovie,
	})
	if err != nil {
		log.Fatalf("get aggregated rating: %v", err)
	}

	if got, want := getAggregatedRatingResp.RatingValue, float64(5); got != want {
		log.Fatalf("rating mismatch: got %v want %v", got, want)
	}

	log.Println("Saving second rating via rating service")
	secondRating := int32(1)
	if _, err := ratingClient.PutRating(ctx, &rpc.PutRatingRequest{
		UserId:      userID,
		RecordId:    m.MovieId,
		RecordType:  recordTypeMovie,
		RatingValue: secondRating,
	}); err != nil {
		log.Fatalf("put rating: %v", err)
	}

	log.Println("Saving new aggregated rating via rating service")
	getAggregatedRatingResp, err = ratingClient.GetAggregatedRating(ctx, &rpc.GetAggregatedRatingRequest{
		RecordId:   m.MovieId,
		RecordType: recordTypeMovie,
	})
	if err != nil {
		log.Fatalf("get aggregated rating: %v", err)
	}

	wantRating := float64((firstRating + secondRating) / 2)
	if got, want := getAggregatedRatingResp.RatingValue, wantRating; got != want {
		log.Fatalf("rating mismatch: got %v want %v", got, want)
	}

	log.Println("Getting updated movie details via movie service")

	getMovieDetailsResp, err = movieClient.GetMovieDetails(ctx, &rpc.GetMovieDetailsRequest{MovieId: m.MovieId})
	if err != nil {
		log.Fatalf("get movie details: %v", err)
	}

	wantMovieDetails.Rating = wantRating
	if diff := cmp.Diff(getMovieDetailsResp.MovieDetails, wantMovieDetails, cmpopts.IgnoreUnexported(rpc.MovieDetails{}, rpc.Metadata{})); diff != "" {
		log.Fatalf("get movie details after update mismatch: %v", err)
	}

	log.Println("Integration test execution successfull")
}

func startMetadataService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting metadata service on ", metadataServiceAddr)
	h := metadatatest.NewTestMetadataGRPCServer()
	l, err := net.Listen("tcp", metadataServiceAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	rpc.RegisterMetadataServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()

	id := discovery.GenerateInstanceID(metadataServiceName)
	if err := registry.Register(ctx, id, metadataServiceName, metadataServiceAddr); err != nil {
		panic(err)
	}

	return srv
}

func startRatingService(ctx context.Context, registry discovery.Registry, cfg *utils.ConfigDatabase) *grpc.Server {
	log.Println("Starting rating service on ", ratingServiceAddr)
	h := ratingtest.NewTestRatingGRPCServer(cfg)
	l, err := net.Listen("tcp", ratingServiceAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	rpc.RegisterRatingServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()

	id := discovery.GenerateInstanceID(ratingServiceName)
	if err := registry.Register(ctx, id, ratingServiceName, ratingServiceAddr); err != nil {
		panic(err)
	}

	return srv
}

func startMovieService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting movie service on ", movieServiceAddr)
	h := movietest.NewTestMovieGRPCServer(registry)
	l, err := net.Listen("tcp", movieServiceAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	rpc.RegisterMovieServiceServer(srv, h)

	go func() {
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()

	id := discovery.GenerateInstanceID(movieServiceName)
	if err := registry.Register(ctx, id, movieServiceName, movieServiceAddr); err != nil {
		panic(err)
	}

	return srv
}
