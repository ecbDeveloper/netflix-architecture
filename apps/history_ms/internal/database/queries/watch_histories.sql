-- name: CreateWatchHistory :one
INSERT INTO watch_histories (id, profile_id, movie_id, episode_id, genre_id, last_position_seconds, is_completed)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetWatchHistory :one
SELECT * FROM watch_histories WHERE id = $1;

-- name: ListWatchHistoryByProfile :many
SELECT * FROM watch_histories WHERE profile_id = $1 ORDER BY watched_at DESC;

-- name: UpdateWatchProgress :one
UPDATE watch_histories
SET last_position_seconds = $2, is_completed = $3, watched_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteWatchHistory :exec
DELETE FROM watch_histories WHERE id = $1;

-- name: GetMostWatchedMovies :many
SELECT movie_id, genre_id, COUNT(*) AS watch_count
FROM watch_histories
WHERE movie_id IS NOT NULL AND is_completed = true
GROUP BY movie_id, genre_id
ORDER BY watch_count DESC
LIMIT $1;

-- name: GetMostWatchedEpisodes :many
SELECT episode_id, genre_id, COUNT(*) AS watch_count
FROM watch_histories
WHERE episode_id IS NOT NULL AND is_completed = true
GROUP BY episode_id, genre_id
ORDER BY watch_count DESC
LIMIT $1;

-- name: GetRecentlyWatchedByProfile :many
SELECT * FROM watch_histories
WHERE profile_id = $1
ORDER BY watched_at DESC
LIMIT $2;
