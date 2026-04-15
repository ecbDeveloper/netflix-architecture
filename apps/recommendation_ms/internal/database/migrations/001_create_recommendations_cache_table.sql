-- Write your migrate up statements here
CREATE TYPE content_type_enum AS ENUM ('MOVIE', 'SERIES');

CREATE TABLE recommendations (
  id UUID PRIMARY KEY,
  profile_id UUID NOT NULL,
  content_id UUID NOT NULL,
  content_type content_type_enum NOT NULL,
  score DOUBLE PRECISION NOT NULL DEFAULT 0,
  reason VARCHAR(50) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
---- create above / drop below ----
DROP TABLE IF EXISTS recommendations;
DROP TYPE IF EXISTS content_type_enum;

