-- name: CreateRating :one
INSERT INTO ratings (
  movie_id,
  record_type,
  user_id,
  value
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetRating :one
SELECT * FROM ratings
WHERE id = $1
ORDER BY id
LIMIT 1;

-- name: ListRatings :many
SELECT * FROM ratings
WHERE movie_id = $1 AND record_type = $2;

-- name: UpdateRating :one
UPDATE ratings
SET
  movie_id = COALESCE(sqlc.narg(movie_id), movie_id),
  record_type = COALESCE(sqlc.narg(record_type), record_type),
  user_id = COALESCE(sqlc.narg(user_id), user_id),
  value = COALESCE(sqlc.narg(value), value)
WHERE
  id = sqlc.arg(id)
RETURNING *;

-- name: DeleteRating :exec
DELETE FROM ratings
WHERE id = $1;