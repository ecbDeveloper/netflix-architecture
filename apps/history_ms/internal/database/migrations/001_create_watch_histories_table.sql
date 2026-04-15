-- Write your migrate up statements here
CREATE TABLE watch_histories (
  id UUID PRIMARY KEY,
  profile_id UUID NOT NULL,
  movie_id UUID,
  episode_id UUID,
  genre_id INT DEFAULT 0,
  watched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  last_position_seconds INT DEFAULT 0,
  is_completed BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_watch_histories_profile_id ON watch_histories(profile_id);
CREATE INDEX idx_watch_histories_genre_id ON watch_histories(genre_id);
---- create above / drop below ----
DROP TABLE IF EXISTS watch_histories;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
