-- Add to users table
ALTER TABLE users ADD COLUMN token_version INTEGER DEFAULT 1;
CREATE INDEX idx_users_token_version ON users(id, token_version);

-- Create refresh tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    last_used_at TIMESTAMP DEFAULT NOW() NOT NULL,
    device_info JSONB DEFAULT '{}' NOT NULL
);

-- Create indexes for refresh tokens
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id, is_revoked, expires_at);
CREATE INDEX idx_refresh_tokens_lookup ON refresh_tokens(token_hash) WHERE is_revoked = FALSE;
CREATE INDEX idx_refresh_tokens_cleanup ON refresh_tokens(expires_at) WHERE is_revoked = FALSE;