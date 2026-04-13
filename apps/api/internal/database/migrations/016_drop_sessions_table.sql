-- Write your migrate up statements here
DROP TABLE IF EXISTS sessions;
---- create above / drop below ----
CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
