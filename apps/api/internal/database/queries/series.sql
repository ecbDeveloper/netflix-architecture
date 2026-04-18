-- name: CreateSerie :one
INSERT INTO series (content_id)
VALUES ($1)
RETURNING *;

-- name: GetSerie :one
SELECT
  c.id, c.title, c.description, c.release_date, c.created_at, c.updated_at, c.maturity_rating, c.genre_id
FROM contents c
JOIN series s ON s.content_id = c.id
WHERE c.id = $1;
