-- Write your migrate up statements here
CREATE TYPE MATURITY_RATING AS ENUM ('L', '10', '12', '14', '16', '18');
CREATE TYPE CONTENT_TYPE    AS ENUM ('MOVIE', 'SERIES');

CREATE TABLE contents (
  id              UUID             PRIMARY KEY,
  title           VARCHAR(255)     NOT NULL UNIQUE,
  content_type    CONTENT_TYPE     NOT NULL,
  genre_id        INT              NOT NULL REFERENCES content_genres(id) ON DELETE RESTRICT ON UPDATE CASCADE,
  description     TEXT             NOT NULL,
  release_date    DATE             NOT NULL,
  maturity_rating MATURITY_RATING  NOT NULL DEFAULT 'L',
  created_at      TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP
);
---- create above / drop below ----
DROP TABLE IF EXISTS contents;
DROP TYPE IF EXISTS CONTENT_TYPE;
DROP TYPE IF EXISTS MATURITY_RATING;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
