package grpc

import (
	"context"
	"errors"
	"main/rating/controller"
	"main/rating/model"
	"main/rpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler defines a gRPC rating API handler.
type Handler struct {
	rpc.UnimplementedRatingServiceServer
	svc *controller.RatingService
}

// New creates a new movie metadata gRPC handler.
func New(svc *controller.RatingService) *Handler {
	return &Handler{
		svc: svc,
	}
}

// GetAggregatedRating returns the aggregated rating for a record.
func (h *Handler) GetAggregatedRating(ctx context.Context, req *rpc.GetAggregatedRatingRequest) (*rpc.GetAggregatedRatingResponse, error) {
	if req == nil || req.RecordId == "" || req.RecordType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or emtpy movie id")
	}

	rating, err := h.svc.GetAggregatedRating(ctx, model.RecordID(req.RecordId), model.RecordType(req.RecordType))
	if err != nil && errors.Is(err, controller.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &rpc.GetAggregatedRatingResponse{
		RatingValue: rating,
	}, nil
}

// PutRating writes a rating for a given record.
func (h *Handler) PutRating(ctx context.Context, req *rpc.PutRatingRequest) (*rpc.PutRatingResponse, error) {
	if req == nil || req.RecordId == "" || req.RecordType == "" || req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty user id or record id")
	}

	if err := h.svc.PutRating(ctx, model.RecordID(req.RecordId), model.RecordType(req.RecordType), &model.Rating{
		RecordID:   req.RecordId,
		RecordType: req.RecordType,
		UserID:     model.UserID(req.UserId),
		Value:      model.RatingValue(req.RatingValue),
	}); err != nil {
		return nil, err
	}

	return &rpc.PutRatingResponse{}, nil
}
