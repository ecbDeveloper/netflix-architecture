-- Write your migrate up statements here
CREATE TABLE series (
  content_id UUID PRIMARY KEY REFERENCES content(id) ON DELETE CASCADE
);
---- create above / drop below ----
DROP TABLE IF EXISTS series;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
