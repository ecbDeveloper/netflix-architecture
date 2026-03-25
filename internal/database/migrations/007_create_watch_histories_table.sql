-- Write your migrate up statements here
CREATE TABLE watch_histories (
  id UUID PRIMARY KEY,
  profile_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
  movie_id UUID REFERENCES movies(id) ON DELETE CASCADE,
  episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
  watched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  last_position_seconds INT DEFAULT 0,
  is_completed BOOLEAN DEFAULT FALSE
);
---- create above / drop below ----
DROP TABLE IF EXISTS watch_histories;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
