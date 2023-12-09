package testutil

import (
	grpchandler "main/rating/handler/grpc"
	"main/rating/repository/memory"
	"main/rating/service"
	"main/rpc"
	"main/util"
)

// NewTestRatingGRPCServer creates a new rating gRPC server to be used in tests.
func NewTestRatingGRPCServer(cfg *util.ConfigDatabase) rpc.RatingServiceServer {
	r := memory.New()
	svc := service.New(r, cfg)
	return grpchandler.New(svc)
}
