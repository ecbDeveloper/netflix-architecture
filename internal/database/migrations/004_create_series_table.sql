-- Write your migrate up statements here
CREATE TABLE series (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  release_date DATE,
  maturity_rating VARCHAR(10)
);
---- create above / drop below ----
IF EXISTS DROP TABLE series;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
