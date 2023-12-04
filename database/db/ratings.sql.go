// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: ratings.sql

package db

import (
	"context"
)

const createRating = `-- name: CreateRating :one
INSERT INTO ratings (
  movie_id,
  record_type,
  user_id,
  value
) VALUES (
  $1, $2, $3, $4
) RETURNING id, movie_id, record_type, user_id, value
`

type CreateRatingParams struct {
	MovieID    string `db:"movie_id" json:"movie_id"`
	RecordType string `db:"record_type" json:"record_type"`
	UserID     string `db:"user_id" json:"user_id"`
	Value      int32  `db:"value" json:"value"`
}

func (q *Queries) CreateRating(ctx context.Context, arg *CreateRatingParams) (*Rating, error) {
	row := q.db.QueryRow(ctx, createRating,
		arg.MovieID,
		arg.RecordType,
		arg.UserID,
		arg.Value,
	)
	var i Rating
	err := row.Scan(
		&i.ID,
		&i.MovieID,
		&i.RecordType,
		&i.UserID,
		&i.Value,
	)
	return &i, err
}

const deleteRating = `-- name: DeleteRating :exec
DELETE FROM ratings
WHERE id = $1
`

func (q *Queries) DeleteRating(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteRating, id)
	return err
}

const getRating = `-- name: GetRating :one
SELECT id, movie_id, record_type, user_id, value FROM ratings
WHERE id = $1
ORDER BY id
LIMIT 1
`

func (q *Queries) GetRating(ctx context.Context, id int64) (*Rating, error) {
	row := q.db.QueryRow(ctx, getRating, id)
	var i Rating
	err := row.Scan(
		&i.ID,
		&i.MovieID,
		&i.RecordType,
		&i.UserID,
		&i.Value,
	)
	return &i, err
}

const listRatings = `-- name: ListRatings :many
SELECT id, movie_id, record_type, user_id, value FROM ratings
WHERE movie_id = $1 AND record_type = $2
`

type ListRatingsParams struct {
	MovieID    string `db:"movie_id" json:"movie_id"`
	RecordType string `db:"record_type" json:"record_type"`
}

func (q *Queries) ListRatings(ctx context.Context, arg *ListRatingsParams) ([]*Rating, error) {
	rows, err := q.db.Query(ctx, listRatings, arg.MovieID, arg.RecordType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*Rating{}
	for rows.Next() {
		var i Rating
		if err := rows.Scan(
			&i.ID,
			&i.MovieID,
			&i.RecordType,
			&i.UserID,
			&i.Value,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateRating = `-- name: UpdateRating :one
UPDATE ratings
SET
  movie_id = COALESCE($1, movie_id),
  record_type = COALESCE($2, record_type),
  user_id = COALESCE($3, user_id),
  value = COALESCE($4, value)
WHERE
  id = $5
RETURNING id, movie_id, record_type, user_id, value
`

type UpdateRatingParams struct {
	MovieID    *string `db:"movie_id" json:"movie_id"`
	RecordType *string `db:"record_type" json:"record_type"`
	UserID     *string `db:"user_id" json:"user_id"`
	Value      *int32  `db:"value" json:"value"`
	ID         int64   `db:"id" json:"id"`
}

func (q *Queries) UpdateRating(ctx context.Context, arg *UpdateRatingParams) (*Rating, error) {
	row := q.db.QueryRow(ctx, updateRating,
		arg.MovieID,
		arg.RecordType,
		arg.UserID,
		arg.Value,
		arg.ID,
	)
	var i Rating
	err := row.Scan(
		&i.ID,
		&i.MovieID,
		&i.RecordType,
		&i.UserID,
		&i.Value,
	)
	return &i, err
}