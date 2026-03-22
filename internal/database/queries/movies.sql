-- name: CreateMovie :one
INSERT INTO movies (id, title, description, duration_minutes, release_date, maturity_rating, content_url)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetMovie :one
SELECT * FROM movies WHERE id = $1;

-- name: ListMovies :many
SELECT * FROM movies ORDER BY release_date DESC;

-- name: UpdateMovie :one
UPDATE movies
SET title = $2, description = $3, duration_minutes = $4, release_date = $5, maturity_rating = $6, content_url = $7
WHERE id = $1
RETURNING *;

-- name: DeleteMovie :exec
DELETE FROM movies WHERE id = $1;
