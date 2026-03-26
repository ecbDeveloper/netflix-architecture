-- Write your migrate up statements here
CREATE TABLE profiles (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  has_parental_controls BOOL NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

  UNIQUE (user_id, name)
);
---- create above / drop below ----
DROP TABLE IF EXISTS profiles;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
