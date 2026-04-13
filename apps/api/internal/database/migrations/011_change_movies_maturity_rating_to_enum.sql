-- Write your migrate up statements here
ALTER TABLE movies
	DROP COLUMN maturity_rating;

CREATE TYPE MATURITY_RATING AS ENUM ('L', '10', '12', '14', '16', '18');

ALTER TABLE movies
	ADD COLUMN maturity_rating MATURITY_RATING NOT NULL DEFAULT 'L';
---- create above / drop below ----
ALTER TABLE movies
	DROP COLUMN maturity_rating;

DROP TYPE MATURITY_RATING;

ALTER TABLE movies
	ADD COLUMN maturity_rating VARCHAR(10);
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
