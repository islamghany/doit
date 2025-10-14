-- name: CreateUser :one
INSERT INTO users (
        id,
        email,
        username,
        password_hash,
        metadata
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
    AND deleted_at IS NULL;
-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1
    AND deleted_at IS NULL;
-- name: ListUsers :many
SELECT *
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
-- name: UpdateUser :one
UPDATE users
SET email = COALESCE(sqlc.narg('email'), email),
    username = COALESCE(sqlc.narg('username'), username),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    metadata = COALESCE(sqlc.narg('metadata'), metadata)
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1
    AND deleted_at IS NULL;
-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;
-- name: VerifyUserEmail :exec
UPDATE users
SET email_verified = TRUE
WHERE id = $1
    AND deleted_at IS NULL;
-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = $1
    AND deleted_at IS NULL;
-- name: CountUsers :one
SELECT COUNT(*)
FROM users
WHERE deleted_at IS NULL;
-- name: SearchUsersByEmail :many
SELECT *
FROM users
WHERE email ILIKE '%' || $1 || '%'
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2;
-- name: BulkUpdateUsersMetadata :exec
UPDATE users
SET metadata = metadata || $2::jsonb
WHERE id = ANY($1::uuid [])
    AND deleted_at IS NULL;