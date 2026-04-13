-- Write your migrate up statements here
ALTER TABLE movies 
  ADD COLUMN genre_id INT NOT NULL DEFAULT 1,
  ADD CONSTRAINT movies_genre_fk 
  FOREIGN KEY (genre_id) 
  REFERENCES content_genres(id)
  ON DELETE RESTRICT
  ON UPDATE CASCADE;
---- create above / drop below ----
ALTER TABLE movies
  DROP CONSTRAINT IF EXISTS movies_genre_fk,
  DROP COLUMN IF EXISTS genre;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
