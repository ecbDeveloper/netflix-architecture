-- Write your migrate up statements here
CREATE TABLE content_genres (
  id SERIAL PRIMARY KEY,
  description VARCHAR(100) NOT NULL
);

INSERT INTO content_genres (description)
VALUES 
  ('Action'),
  ('Comedy'),
  ('Drama'),
  ('Horror'),
  ('Sci-Fi'),
  ('Documentary'),
  ('Animation'),
  ('Thriller'),
  ('Romance'),
  ('Fantasy'),
  ('Mystery'),
  ('Crime'),
  ('Family'),
  ('Adventure'),
  ('War'),
  ('Western'),
  ('Musical'),
  ('Biography'),
  ('History'),
  ('Sport');
---- create above / drop below ----
DROP TABLE IF EXISTS content_genres;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
