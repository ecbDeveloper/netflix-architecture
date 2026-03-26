-- Write your migrate up statements here
CREATE TABLE reviews (
  id SERIAL PRIMARY KEY,
  profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  movie_id UUID REFERENCES movies(id) ON DELETE CASCADE,
  episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
  rating INT NOT NULL,
  comment TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
---- create above / drop below ----
DROP TABLE IF EXISTS reviews;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
