-- name: CreateSerie :one
INSERT INTO series (title, description, release_date, maturity_rating)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSerie :one
SELECT * FROM series WHERE id = $1;

-- name: ListSeries :many
SELECT * FROM series ORDER BY title;

-- name: UpdateSerie :one
UPDATE series
SET title = $2, description = $3, release_date = $4, maturity_rating = $5
WHERE id = $1
RETURNING *;

-- name: DeleteSerie :exec
DELETE FROM series WHERE id = $1;
