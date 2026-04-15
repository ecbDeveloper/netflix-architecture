-- name: CreateRecommendation :one
INSERT INTO recommendations (id, profile_id, content_id, content_type, score, reason)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListRecommendationsByProfile :many
SELECT * FROM recommendations
WHERE profile_id = $1
ORDER BY score DESC
LIMIT $2;

-- name: DeleteRecommendationsByProfile :exec
DELETE FROM recommendations WHERE profile_id = $1;

-- name: CountRecommendationsByProfile :one
SELECT COUNT(*) FROM recommendations WHERE profile_id = $1;
