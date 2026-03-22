-- name: CreateUser :one
INSERT INTO users (id, email, name, cpf, password, salt)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY name;

-- name: UpdateUser :one
UPDATE users 
SET email = $2, name = $3, cpf = $4, password = $5, salt = $6, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
