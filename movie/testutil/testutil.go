package testutil

import (
	"main/discovery"
	metadatagateway "main/movie/gateway/metadata/grpc"
	ratinggateway "main/movie/gateway/rating/grpc"
	grpchandler "main/movie/handler/grpc"
	"main/movie/service"
	"main/rpc"
)

// NewTestMovieGRPCServer creates a new movie gRPC server to be used in tests.
func NewTestMovieGRPCServer(registry discovery.Registry) rpc.MovieServiceServer {
	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	ctrl := service.New(ratingGateway, metadataGateway)
	return grpchandler.New(ctrl)
}
