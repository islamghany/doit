-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    id, user_id, token_hash, expires_at, device_info
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1 
  AND is_revoked = FALSE 
  AND expires_at > NOW()
LIMIT 1;

-- name: GetRefreshTokenIncludingRevoked :one
-- For security checks (detect reuse)
SELECT * FROM refresh_tokens
WHERE token_hash = $1
LIMIT 1;


-- name: UpdateRefreshTokenUsage :exec
UPDATE refresh_tokens 
SET last_used_at = NOW()
WHERE id = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens 
SET is_revoked = TRUE
WHERE token_hash = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens 
SET is_revoked = TRUE
WHERE user_id = $1 AND is_revoked = FALSE;

-- name: GetUserActiveRefreshTokens :many
SELECT id, created_at, last_used_at, device_info
FROM refresh_tokens
WHERE user_id = $1 
  AND is_revoked = FALSE 
  AND expires_at > NOW()
ORDER BY last_used_at DESC;

-- name: IncrementUserTokenVersion :one
UPDATE users 
SET token_version = token_version + 1
WHERE id = $1
RETURNING token_version;

-- name: GetUserTokenVersion :one
SELECT token_version FROM users WHERE id = $1;

-- name: CleanupExpiredTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at < NOW() - INTERVAL '30 days';