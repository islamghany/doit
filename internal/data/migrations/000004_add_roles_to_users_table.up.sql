-- Add role column to users table
ALTER TABLE users
ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';
-- Create index for role-based queries
CREATE INDEX idx_users_role ON users(role);
-- Add check constraint
ALTER TABLE users
ADD CONSTRAINT check_user_role CHECK (role IN ('user', 'admin', 'moderator'));