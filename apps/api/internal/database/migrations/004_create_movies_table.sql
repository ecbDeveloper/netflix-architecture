-- Write your migrate up statements here
CREATE TABLE movies (
  content_id UUID PRIMARY KEY REFERENCES content(id) ON DELETE CASCADE,
  duration_minutes INT NOT NULL,
  content_url TEXT NOT NULL
);
---- create above / drop below ----
DROP TABLE IF EXISTS movies;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
