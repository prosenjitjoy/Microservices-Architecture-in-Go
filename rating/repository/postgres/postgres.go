package postgres

import (
	"context"
	"go.opentelemetry.io/otel"
	"main/database/db"
	"main/rating/model"
	"main/rating/repository"
)

const tracerID = "rating-repository-postgres"

// Repository defines a PostgreSQL-based movie metadata repository.
type Repository struct {
	db db.Store
}

// New creates a new PostgreSQL-based repository.
func New(store db.Store) *Repository {
	return &Repository{
		db: store,
	}
}

// Get retrieves all ratings for a given record.
func (r *Repository) Get(ctx context.Context, movieId model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/GET")
	defer span.End()

	ratings, err := r.db.ListRatings(ctx, &db.ListRatingsParams{
		MovieID:    string(movieId),
		RecordType: string(recordType),
	})
	if err != nil {
		return nil, err
	}

	var res []model.Rating
	for _, rating := range ratings {
		res = append(res, model.Rating{
			UserID: model.UserID(rating.UserID),
			Value:  model.RatingValue(rating.Value),
		})
	}

	if len(res) == 0 {
		return nil, repository.ErrNotFound
	}

	return res, nil
}

// Put adds a rating for a given record.
func (r *Repository) Put(ctx context.Context, movieId model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/PUT")
	defer span.End()

	_, err := r.db.CreateRating(ctx, &db.CreateRatingParams{
		MovieID:    string(movieId),
		RecordType: string(recordType),
		UserID:     string(rating.UserID),
		Value:      int32(rating.Value),
	})

	return err
}
