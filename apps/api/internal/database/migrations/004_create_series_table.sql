-- Write your migrate up statements here
CREATE TABLE series (
  id UUID PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  release_date DATE NOT NULL,
  maturity_rating VARCHAR(10) NOT NULL,
  
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP NOT NULL
);
---- create above / drop below ----
DROP TABLE IF EXISTS series;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
