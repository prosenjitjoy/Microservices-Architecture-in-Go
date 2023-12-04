package postgres

import (
	"context"
	"main/database/db"
	"main/metadata/model"
	"main/metadata/repository"

	"github.com/jackc/pgx/v5"
)

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

// Get retrieves movie metadata for by movie id.
func (r *Repository) Get(ctx context.Context, id string) (*model.Metadata, error) {
	movie, err := r.db.GetMovie(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &model.Metadata{
		ID:          movie.ID,
		Title:       movie.Title,
		Description: movie.Description,
		Director:    movie.Description,
	}, nil
}

// Put adds movie metadata for a given movie id.
func (r *Repository) Put(ctx context.Context, id string, metadata *model.Metadata) error {
	_, err := r.db.CreateMovie(ctx, &db.CreateMovieParams{
		ID:          id,
		Title:       metadata.Title,
		Description: metadata.Description,
		Director:    metadata.Director,
	})
	return err
}

// CREATE TABLE IF NOT EXISTS "movies" (
//   "id" text PRIMARY KEY,
//   "title" text UNIQUE NOT NULL,
//   "description" text NOT NULL,
//   "director" text NOT NULL
// );

// CREATE TABLE IF NOT EXISTS "ratings" (
//   "id" bigserial PRIMARY KEY,
//   "movie_id" text NOT NULL,
//   "record_type" text NOT NULL,
//   "user_id" text NOT NULL,
//   "value" integer NOT NULL
// );
