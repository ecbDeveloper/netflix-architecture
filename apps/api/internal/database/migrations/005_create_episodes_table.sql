-- Write your migrate up statements here
CREATE TABLE episodes (
  id UUID PRIMARY KEY,
  series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE ON UPDATE CASCADE,
  season INT NOT NULL,
  episode_number INT NOT NULL,
  title VARCHAR(255) NOT NULL,
  duration_minutes INT NOT NULL,
  content_url TEXT NOT NULL,

  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP NOT NULL,

  UNIQUE (series_id, season, episode_number)
);
---- create above / drop below ----
DROP TABLE IF EXISTS episodes;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
