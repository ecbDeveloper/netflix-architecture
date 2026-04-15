-- name: CreateEpisode :one
INSERT INTO episodes (id, series_id, season, episode_number, title, duration_minutes, content_url)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetEpisode :one
SELECT * FROM episodes WHERE id = $1;

-- name: ListEpisodesBySerie :many
SELECT * FROM episodes WHERE series_id = $1 ORDER BY season, episode_number;

-- name: UpdateEpisode :one
UPDATE episodes
SET season = $2, episode_number = $3, title = $4, duration_minutes = $5, content_url = $6
WHERE id = $1
RETURNING *;

-- name: DeleteEpisode :exec
DELETE FROM episodes WHERE id = $1;
