-- Remove from users table
ALTER TABLE users DROP COLUMN token_version;
DROP INDEX idx_users_token_version;

-- Remove refresh tokens table
DROP TABLE refresh_tokens;

-- Remove indexes for refresh tokens
DROP INDEX idx_refresh_tokens_user;
DROP INDEX idx_refresh_tokens_lookup;
DROP INDEX idx_refresh_tokens_cleanup;