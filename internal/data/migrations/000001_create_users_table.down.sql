-- Drop users table
DROP TABLE users;
-- Drop indexes for users
DROP INDEX idx_users_email;
DROP INDEX idx_users_username;
DROP INDEX idx_users_created_at;
DROP INDEX idx_users_metadata_gin;
-- Drop triggers for updated_at
DROP TRIGGER update_users_updated_at;
DROP FUNCTION update_updated_at_column;
-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS "pg_trgm";