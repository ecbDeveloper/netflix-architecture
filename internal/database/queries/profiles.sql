-- name: CreateProfile :one
INSERT INTO profiles (id, user_id, name, has_parental_controls)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProfile :one
SELECT * FROM profiles WHERE id = $1;

-- name: ListProfilesByUser :many
SELECT * FROM profiles WHERE user_id = $1;

-- name: UpdateProfile :one
UPDATE profiles 
SET name = $2, has_parental_controls = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteProfile :exec
DELETE FROM profiles WHERE id = $1;
