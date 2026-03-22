-- name: CreateReview :one
INSERT INTO reviews (profile_id, movie_id, episode_id, rating, comment)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetReview :one
SELECT * FROM reviews WHERE id = $1;

-- name: ListReviewsByMovie :many
SELECT * FROM reviews WHERE movie_id = $1;

-- name: ListReviewsByEpisode :many
SELECT * FROM reviews WHERE episode_id = $1;

-- name: DeleteReview :exec
DELETE FROM reviews WHERE id = $1;
