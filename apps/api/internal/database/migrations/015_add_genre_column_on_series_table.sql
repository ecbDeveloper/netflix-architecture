-- Write your migrate up statements here
ALTER TABLE series 
  ADD COLUMN genre_id INT NOT NULL DEFAULT 1,
  ADD CONSTRAINT series_genre_fk 
  FOREIGN KEY (genre_id) 
  REFERENCES content_genres(id)
  ON DELETE RESTRICT
  ON UPDATE CASCADE;
---- create above / drop below ----
ALTER TABLE series
  DROP CONSTRAINT IF EXISTS series_genre_fk,
  DROP COLUMN IF EXISTS genre;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
