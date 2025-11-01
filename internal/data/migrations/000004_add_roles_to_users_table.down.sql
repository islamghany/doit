ALTER TABLE users DROP CONSTRAINT check_user_role;
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN role;