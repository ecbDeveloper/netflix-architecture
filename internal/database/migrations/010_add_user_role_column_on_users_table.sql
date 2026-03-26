-- Write your migrate up statements here
ALTER TABLE users 
ADD COLUMN role INT NOT NULL DEFAULT 2,
ADD CONSTRAINT users_roles_user_fk FOREIGN KEY (role)
  REFERENCES users_roles(id);
---- create above / drop below ----
ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_roles_user_fk,
  DROP COLUMN IF EXISTS role; 
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
