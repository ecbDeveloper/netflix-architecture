-- name: CreateReview :one
INSERT INTO reviews (profile_id, movie_id, episode_id, rating, comment)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetReview :one
SELECT * FROM reviews WHERE id = $1;

-- name: ListReviewsByProfile :many
SELECT * FROM reviews WHERE profile_id = $1 ORDER BY created_at DESC;

-- name: ListReviewsByMovie :many
SELECT * FROM reviews WHERE movie_id = $1;

-- name: ListReviewsByEpisode :many
SELECT * FROM reviews WHERE episode_id = $1;

-- name: UpdateReview :one
UPDATE reviews
SET rating = $2, comment = $3
WHERE id = $1
RETURNING *;

-- name: DeleteReview :exec
DELETE FROM reviews WHERE id = $1;
