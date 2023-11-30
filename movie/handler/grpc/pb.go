package grpc

import (
	"context"
	"errors"
	"main/metadata/model"
	"main/movie/controller"
	"main/rpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler defines a movie gRPC handler.
type Handler struct {
	rpc.UnimplementedMovieServiceServer
	svc *controller.MovieService
}

// New creates a new movie gRPC handler.
func New(svc *controller.MovieService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) GetMovieDetails(ctx context.Context, req *rpc.GetMovieDetailsRequest) (*rpc.GetMovieDetailsResponse, error) {
	if req == nil || req.MovieId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty movie id")
	}

	m, err := h.svc.Get(ctx, req.MovieId)
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &rpc.GetMovieDetailsResponse{
		MovieDetails: &rpc.MovieDetails{
			Rating:   *m.Rating,
			Metadata: model.MetadataToProto(&m.Metadata),
		},
	}, nil
}
