package testutil

import (
	grpchandler "main/metadata/handler/grpc"
	"main/metadata/repository/memory"
	"main/metadata/service"
	"main/rpc"
)

// NewTestMetadataGRPCServer creates a new metadata gRPC server to be used in tests.
func NewTestMetadataGRPCServer() rpc.MetadataServiceServer {
	r := memory.New()
	svc := service.New(r)
	return grpchandler.New(svc)
}
