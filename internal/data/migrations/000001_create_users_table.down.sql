-- Drop triggers for updated_at (before dropping table)
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
-- Drop users table (this also drops all indexes automatically)
DROP TABLE IF EXISTS users;
-- Drop function (only if no other tables use it)
DROP FUNCTION IF EXISTS update_updated_at_column();
-- Drop extensions (careful: might be used by other databases/schemas)
DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "uuid-ossp";