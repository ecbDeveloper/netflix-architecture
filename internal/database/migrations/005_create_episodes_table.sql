-- Write your migrate up statements here
CREATE TABLE episodes (
  id UUID PRIMARY KEY,
  serie_id INT REFERENCES series(id) ON DELETE CASCADE,
  season INT NOT NULL,
  episode_number INT NOT NULL,
  title VARCHAR(255) NOT NULL,
  duration_minutes INT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMPTZ NOT NULL,

  UNIQUE (series_id, season, episode_number)
);
---- create above / drop below ----
IF EXISTS DROP TABLE episodes;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
