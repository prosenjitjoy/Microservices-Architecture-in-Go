package main

import (
	"context"
	"flag"
	"log/slog"
	"main/discovery"
	"main/discovery/memory"
	"main/rpc"
	"main/util"
	"net"
	"os"

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
	cfg := util.LoadConfig(config)

	if cfg.Environment == "dev" {
		var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
		slog.SetDefault(logger)
	} else {
		var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)
	}

	slog.Info("Starting the integration test")

	ctx := context.Background()
	registry := memory.NewRegistry()

	slog.Info("Setting up service handlers and clients")

	metadataSrv := startMetadataService(ctx, registry)
	defer metadataSrv.GracefulStop()
	ratingSrv := startRatingService(ctx, registry, cfg)
	defer ratingSrv.GracefulStop()
	movieSrv := startMovieService(ctx, registry)
	defer movieSrv.GracefulStop()

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	metadataConn, err := grpc.Dial(metadataServiceAddr, opts)
	if err != nil {
		slog.Error("failed to create metadata gRPC client:", slog.String("error", err.Error()))
		return
	}
	defer metadataConn.Close()
	metadataClient := rpc.NewMetadataServiceClient(metadataConn)

	ratingConn, err := grpc.Dial(ratingServiceAddr, opts)
	if err != nil {
		slog.Error("failed to create rating gRPC client:", slog.String("error", err.Error()))
		return
	}
	defer ratingConn.Close()
	ratingClient := rpc.NewRatingServiceClient(ratingConn)

	movieConn, err := grpc.Dial(movieServiceAddr, opts)
	if err != nil {
		slog.Error("failed to create movie gRPC client:", slog.String("error", err.Error()))
		return
	}
	defer movieConn.Close()
	movieClient := rpc.NewMovieServiceClient(movieConn)

	slog.Info("Saving test metadata via metadata service")

	m := &rpc.Metadata{
		MovieId:     "the-movie",
		Title:       "The Movie",
		Description: "The Movie, the one and only",
		Director:    "Mr. D",
	}

	if _, err := metadataClient.PutMetadata(ctx, &rpc.PutMetadataRequest{
		Metadata: m,
	}); err != nil {
		slog.Error("failed to put metadata:", slog.String("error", err.Error()))
		return
	}

	slog.Info("Retrieving test metadata via metadata service")

	getMetadataResp, err := metadataClient.GetMetadata(ctx, &rpc.GetMetadataRequest{MovieId: m.MovieId})
	if err != nil {
		slog.Error("get metadata:", slog.String("error", err.Error()))
		return
	}
	if diff := cmp.Diff(getMetadataResp.Metadata, m, cmpopts.IgnoreUnexported(rpc.Metadata{})); diff != "" {
		slog.Error("get metadata after put mismatch:", slog.String("diff", diff))
		return
	}

	slog.Info("Getting movie details via movie service")
	wantMovieDetails := &rpc.MovieDetails{
		Metadata: m,
	}

	getMovieDetailsResp, err := movieClient.GetMovieDetails(ctx, &rpc.GetMovieDetailsRequest{MovieId: m.MovieId})
	if err != nil {
		slog.Error("geet movie details", slog.String("error", err.Error()))
		return
	}

	if diff := cmp.Diff(getMovieDetailsResp.MovieDetails, wantMovieDetails, cmpopts.IgnoreUnexported(rpc.MovieDetails{}, rpc.Metadata{})); diff != "" {
		slog.Error("get movie details after put mismatch:", slog.String("error", err.Error()))
		return
	}

	slog.Info("Saving first rating via rating service")

	const userID = "user0"
	const recordTypeMovie = "movie"
	firstRating := int32(5)
	if _, err := ratingClient.PutRating(ctx, &rpc.PutRatingRequest{
		UserId:      userID,
		RecordId:    m.MovieId,
		RecordType:  recordTypeMovie,
		RatingValue: firstRating,
	}); err != nil {
		slog.Error("put rating", slog.String("error", err.Error()))
		return
	}

	slog.Info("Retrieving initial aggregated rating via rating service")

	getAggregatedRatingResp, err := ratingClient.GetAggregatedRating(ctx, &rpc.GetAggregatedRatingRequest{
		RecordId:   m.MovieId,
		RecordType: recordTypeMovie,
	})
	if err != nil {
		slog.Error("get aggregated rating:", slog.String("error", err.Error()))
		return
	}

	if got, want := getAggregatedRatingResp.RatingValue, float64(5); got != want {
		slog.Error("rating mismatch:", slog.Float64("got", got), slog.Float64("want", want))
		return
	}

	slog.Info("Saving second rating via rating service")
	secondRating := int32(1)
	if _, err := ratingClient.PutRating(ctx, &rpc.PutRatingRequest{
		UserId:      userID,
		RecordId:    m.MovieId,
		RecordType:  recordTypeMovie,
		RatingValue: secondRating,
	}); err != nil {
		slog.Error("put rating:", slog.String("error", err.Error()))
		return
	}

	slog.Info("Saving new aggregated rating via rating service")
	getAggregatedRatingResp, err = ratingClient.GetAggregatedRating(ctx, &rpc.GetAggregatedRatingRequest{
		RecordId:   m.MovieId,
		RecordType: recordTypeMovie,
	})
	if err != nil {
		slog.Error("get aggregated rating:", slog.String("error", err.Error()))
		return
	}

	wantRating := float64((firstRating + secondRating) / 2)
	if got, want := getAggregatedRatingResp.RatingValue, wantRating; got != want {
		slog.Error("rating mismatch:", slog.Float64("got", got), slog.Float64("want", want))
		return
	}

	slog.Info("Getting updated movie details via movie service")

	getMovieDetailsResp, err = movieClient.GetMovieDetails(ctx, &rpc.GetMovieDetailsRequest{MovieId: m.MovieId})
	if err != nil {
		slog.Error("get movie details:", slog.String("error", err.Error()))
		return
	}

	wantMovieDetails.Rating = wantRating
	if diff := cmp.Diff(getMovieDetailsResp.MovieDetails, wantMovieDetails, cmpopts.IgnoreUnexported(rpc.MovieDetails{}, rpc.Metadata{})); diff != "" {
		slog.Error("get movie details after update mismatch:", slog.String("error", err.Error()))
		return
	}

	slog.Info("Integration test execution successfull")
}

func startMetadataService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	slog.Info("Starting metadata service on ", slog.String("address", metadataServiceAddr))
	h := metadatatest.NewTestMetadataGRPCServer()
	l, err := net.Listen("tcp", metadataServiceAddr)
	if err != nil {
		slog.Error("failed to listen:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	srv := grpc.NewServer()
	rpc.RegisterMetadataServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			slog.Error("failed to serve:", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	id := discovery.GenerateInstanceID(metadataServiceName)
	if err := registry.Register(ctx, id, metadataServiceName, metadataServiceAddr); err != nil {
		slog.Error("failed to resister:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	return srv
}

func startRatingService(ctx context.Context, registry discovery.Registry, cfg *util.ConfigDatabase) *grpc.Server {
	slog.Info("Starting rating service on ", slog.String("address", ratingServiceAddr))
	h := ratingtest.NewTestRatingGRPCServer(cfg)
	l, err := net.Listen("tcp", ratingServiceAddr)
	if err != nil {
		slog.Error("failed to listen:", slog.String("error", err.Error()))
		os.Exit(1)
	}
	srv := grpc.NewServer()
	rpc.RegisterRatingServiceServer(srv, h)
	go func() {
		if err := srv.Serve(l); err != nil {
			slog.Error("failed to serve:", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	id := discovery.GenerateInstanceID(ratingServiceName)
	if err := registry.Register(ctx, id, ratingServiceName, ratingServiceAddr); err != nil {
		slog.Error("failed to resister:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	return srv
}

func startMovieService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	slog.Info("Starting movie service on ", slog.String("address", movieServiceAddr))
	h := movietest.NewTestMovieGRPCServer(registry)
	l, err := net.Listen("tcp", movieServiceAddr)
	if err != nil {
		slog.Error("failed to listen:", slog.String("error", err.Error()))
		os.Exit(1)
	}
	srv := grpc.NewServer()
	rpc.RegisterMovieServiceServer(srv, h)

	go func() {
		if err := srv.Serve(l); err != nil {
			slog.Error("failed to serve:", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	id := discovery.GenerateInstanceID(movieServiceName)
	if err := registry.Register(ctx, id, movieServiceName, movieServiceAddr); err != nil {
		slog.Error("failed to resister:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	return srv
}
