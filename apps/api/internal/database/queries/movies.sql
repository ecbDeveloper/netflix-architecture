-- name: CreateMovie :one
INSERT INTO movies (
  content_id
) VALUES ($1)
RETURNING *;

-- name: GetMovie :one
SELECT
  c.id, c.title, c.description, m.duration_seconds, c.release_date,
  m.content_url, c.created_at, c.updated_at, c.maturity_rating, c.genre_id, m.status
FROM contents c
JOIN movies m ON m.content_id = c.id
WHERE c.id = $1;

-- name: UpdateMovie :one
UPDATE movies
SET duration_seconds = $2, content_url = $3, status = $4
WHERE content_id = $1
RETURNING *;
