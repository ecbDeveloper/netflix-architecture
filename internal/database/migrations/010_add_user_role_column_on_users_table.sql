-- Write your migrate up statements here
ALTER TABLE users 
ADD COLUMN role_id INT NOT NULL DEFAULT 2,
ADD CONSTRAINT users_roles_user_fk FOREIGN KEY (role_id)
  REFERENCES users_roles(id)
  ON UPDATE CASCADE;
---- create above / drop below ----
ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_roles_user_fk,
  DROP COLUMN IF EXISTS role_id; 
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
