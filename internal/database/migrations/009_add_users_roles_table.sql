-- Write your migrate up statements here
CREATE TABLE users_roles (
  id SERIAL PRIMARY KEY,
  role VARCHAR(20) UNIQUE
);

INSERT INTO users_roles (role)
VALUES ('admin'), ('member');
---- create above / drop below ----
DROP TABLE IF EXISTS  users_roles;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
