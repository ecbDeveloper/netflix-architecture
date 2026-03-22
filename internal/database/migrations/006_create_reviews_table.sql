-- Write your migrate up statements here
CREATE TABLE reviews (
  id SERIAL PRIMARY KEY,
  profile_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
  movie_id UUID REFERENCES movies(id) ON DELETE CASCADE,
  episode_id UUID REFERENCES episodies(id) ON DELETE CASCADE,
  rating INT NOT NULL,
  comment TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
---- create above / drop below ----
IF EXISTS DROP TABLE reviews;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
