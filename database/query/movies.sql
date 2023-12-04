-- name: CreateMovie :one
INSERT INTO movies (
  id,
  title,
  description,
  director
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetMovie :one
SELECT * FROM movies
WHERE id = $1
ORDER BY id
LIMIT 1;

-- name: ListMovies :many
SELECT * FROM movies
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateMovie :one
UPDATE movies
SET
  title = COALESCE(sqlc.narg(title), title),
  description = COALESCE(sqlc.narg(description), description),
  director = COALESCE(sqlc.narg(director), director)
WHERE
  id = sqlc.arg(id)
RETURNING *;

-- name: DeleteMovie :exec
DELETE FROM movies
WHERE id = $1;