-- Write your migrate up statements here
CREATE TYPE CONTENT_STATUS AS ENUM ('PENDING', 'PROCESSED');

CREATE TABLE movies (
  content_id       UUID PRIMARY KEY REFERENCES contents(id) ON DELETE CASCADE,
  duration_minutes INT  NOT NULL,
  content_url      TEXT,
  status           CONTENT_STATUS NOT NULL DEFAULT 'PENDING',
  created_at       TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);
---- create above / drop below ----
DROP TABLE IF EXISTS movies;
DROP TYPE IF EXISTS CONTENT_STATUS;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
