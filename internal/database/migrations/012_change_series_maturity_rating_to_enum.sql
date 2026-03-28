-- Write your migrate up statements here
ALTER TABLE	series 
	DROP COLUMN maturity_rating;

ALTER TABLE series
	ADD COLUMN maturity_rating MATURITY_RATING NOT NULL DEFAULT 'L';
---- create above / drop below ----
ALTER TABLE series
	DROP COLUMN maturity_rating;

ALTER TABLE series
	ADD COLUMN maturity_rating VARCHAR(10);
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
