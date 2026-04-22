-- name: CreateContent :exec
INSERT INTO contents (
  id, 
  title, 
  content_type, 
  description, 
  release_date, 
  maturity_rating
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListContents :many
SELECT * FROM contents;

-- name: ListKidsContents :many
SELECT * FROM contents
WHERE maturity_rating = 'L';

-- name: ListContentsByType :many
SELECT * FROM contents
WHERE content_type = $1;

-- name: ListContentsByGenre :many
SELECT * FROM contents
WHERE genre_id = $1;

-- name: DeleteContent :exec
DELETE FROM contents
WHERE id = $1;

-- name: UpdateContent :one
UPDATE contents SET
  title = $2, 
	genre_id = $3,
  description = $4, 
  release_date = $5, 
  maturity_rating = $6
WHERE id = $1
RETURNING *;

-- name: GetContent :one
SELECT *
FROM contents
WHERE id = $1;
