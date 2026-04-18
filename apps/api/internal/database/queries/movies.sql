-- name: CreateMovie :one
INSERT INTO movies (
  content_id, 
  duration_minutes, 
  content_url
) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMovie :one
SELECT
  c.id, c.title, c.description, m.duration_minutes, c.release_date,
  m.content_url, c.created_at, c.updated_at, c.maturity_rating, c.genre_id
FROM contents c
JOIN movies m ON m.content_id = c.id
WHERE c.id = $1;

-- name: UpdateMovie :one
UPDATE movies
SET duration_minutes = $2, content_url = $3
WHERE content_id = $1
RETURNING *;
