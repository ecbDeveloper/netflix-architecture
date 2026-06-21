-- name: UpsertMovieWatchHistory :one
INSERT INTO watch_histories (
  id, 
  profile_id, 
  movie_id, 
  genre_id, 
  last_position_seconds, 
  is_completed
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (profile_id, movie_id) 
DO UPDATE SET 
  last_position_seconds = EXCLUDED.last_position_seconds,
  is_completed = EXCLUDED.is_completed,
  watched_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: UpsertEpisodeWatchHistory :one
INSERT INTO watch_histories (
  id, 
  profile_id, 
  episode_id, 
  genre_id, 
  last_position_seconds, 
  is_completed
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (profile_id, episode_id) 
DO UPDATE SET 
  last_position_seconds = EXCLUDED.last_position_seconds,
  is_completed = EXCLUDED.is_completed,
  watched_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetWatchHistory :one
SELECT * FROM watch_histories WHERE id = $1;

-- name: ListWatchHistoryByProfile :many
SELECT * FROM watch_histories WHERE profile_id = $1 ORDER BY watched_at DESC;

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
