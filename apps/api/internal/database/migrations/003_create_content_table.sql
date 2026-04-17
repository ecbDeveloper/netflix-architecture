-- Write your migrate up statements here
CREATE TYPE MATURITY_RATING AS ENUM ('L', '10', '12', '14', '16', '18');
CREATE TYPE CONTENT_TYPE AS ENUM ('MOVIE', 'SERIES');

CREATE TABLE content (
  id UUID PRIMARY KEY,
  title VARCHAR(255) NOT NULL UNIQUE,
  content_type CONTENT_TYPE NOT NULL,
  description TEXT NOT NULL,
  release_date DATE NOT NULL,
  maturity_rating MATURITY_RATING NOT NULL DEFAULT 'L',

  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP NOT NULL
);
---- create above / drop below ----
DROP TABLE IF EXISTS content;
DROP TYPE IF EXISTS CONTENT_TYPE;
DROP TYPE IF EXISTS MATURITY_RATING;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
