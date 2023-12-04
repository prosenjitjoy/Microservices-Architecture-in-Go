package grpc

import (
	"context"
	"main/discovery"
	"main/metadata/model"
	"main/rpc"
	"main/utils"
)

// Gateway defines a movie metadata gRPC gateway.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new gRPC gateway for a movie metadata service.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{
		registry: registry,
	}
}

// Get returns movie metadata by a movie id.
func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	conn, err := utils.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := rpc.NewMetadataServiceClient(conn)
	resp, err := client.GetMetadata(ctx, &rpc.GetMetadataRequest{
		MovieId: id,
	})
	if err != nil {
		return nil, err
	}

	return model.MetadataFromProto(resp.Metadata), nil
}
