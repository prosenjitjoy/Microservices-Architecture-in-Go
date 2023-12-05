package grpc

import (
	"context"
	"errors"
	"main/metadata/model"
	"main/metadata/service"
	"main/rpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler defines a movie metadata gRPC handler.
type Handler struct {
	rpc.UnimplementedMetadataServiceServer
	svc *service.MetadataService
}

// New creates a new movie metadata gRPC handler.
func New(svc *service.MetadataService) *Handler {
	return &Handler{
		svc: svc,
	}
}

// GetMetadata returns movie metadata by id.
func (h *Handler) GetMetadata(ctx context.Context, req *rpc.GetMetadataRequest) (*rpc.GetMetadataResponse, error) {
	if req == nil || req.MovieId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty movie id")
	}

	m, err := h.svc.GetMetadata(ctx, req.MovieId)
	if err != nil && errors.Is(err, service.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &rpc.GetMetadataResponse{
		Metadata: model.MetadataToProto(m),
	}, nil
}

// PutMetadata insert a movie metadata.
func (h *Handler) PutMetadata(ctx context.Context, req *rpc.PutMetadataRequest) (*rpc.PutMetadataResponse, error) {
	if req == nil || req.Metadata.MovieId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty movie id")
	}

	id := req.Metadata.MovieId
	metadata := model.MetadataFromProto(req.Metadata)
	err := h.svc.PutMetadata(ctx, id, metadata)
	if err != nil && errors.Is(err, service.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &rpc.PutMetadataResponse{}, nil
}
