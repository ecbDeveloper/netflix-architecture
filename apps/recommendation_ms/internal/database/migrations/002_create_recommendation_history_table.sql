-- Write your migrate up statements here
CREATE TABLE recommendation_history (
    id VARCHAR(255) PRIMARY KEY,
    profile_id VARCHAR(255),
    movie_id VARCHAR(255),
    episode_id VARCHAR(255),
    genre_id INT,
    watched_at BIGINT,
    is_completed BOOLEAN
);

---- create above / drop below ----
DROP TABLE IF EXISTS recommendation_history;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
