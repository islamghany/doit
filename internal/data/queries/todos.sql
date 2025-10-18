-- name: CreateTodo :one
INSERT INTO todos (
        id,
        user_id,
        title,
        description,
        status,
        priority,
        tags,
        metadata,
        due_date
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;
-- name: GetTodoByID :one
SELECT *
FROM todos
WHERE id = $1;
-- name: GetTodoByIDForUpdate :one
-- Use FOR UPDATE to lock the row for updates (prevents race conditions)
SELECT *
FROM todos
WHERE id = $1 FOR
UPDATE;
-- name: ListTodosByUser :many
SELECT *
FROM todos
WHERE user_id = $1
ORDER BY CASE
        priority
        WHEN 'urgent' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END,
    created_at DESC
LIMIT $2 OFFSET $3;
-- name: ListTodosByUserAndStatus :many
SELECT *
FROM todos
WHERE user_id = $1
    AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;
-- name: UpdateTodo :one
UPDATE todos
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    status = COALESCE(sqlc.narg('status'), status),
    priority = COALESCE(sqlc.narg('priority'), priority),
    tags = COALESCE(sqlc.narg('tags'), tags),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    due_date = COALESCE(sqlc.narg('due_date'), due_date)
WHERE id = sqlc.arg('id')
RETURNING *;
-- name: CompleteTodo :one
UPDATE todos
SET status = 'completed',
    completed_at = NOW()
WHERE id = $1
    AND status != 'completed'
RETURNING *;
-- name: CountUserTodos :one
SELECT COUNT(*)
FROM todos
WHERE user_id = $1;
-- name: CountUserTodosByStatus :one
SELECT COUNT(*)
FROM todos
WHERE user_id = $1
    AND status = $2;
-- name: GetOverdueTodos :many
SELECT t.*,
    u.email as user_email
FROM todos t
    JOIN users u ON t.user_id = u.id
WHERE t.due_date < NOW()
    AND t.status != 'completed'
ORDER BY t.due_date ASC
LIMIT $1;
-- name: SearchTodosByTitle :many
SELECT *
FROM todos
WHERE user_id = $1
    AND title ILIKE '%' || $2 || '%'
ORDER BY created_at DESC
LIMIT $3;
-- name: GetTodosByTags :many
SELECT *
FROM todos
WHERE user_id = $1
    AND tags && $2::text []
ORDER BY created_at DESC;
-- name: BulkUpdateTodoStatus :exec
UPDATE todos
SET status = $2
WHERE id = ANY($1::uuid []);
-- name: GetTodoStats :one
SELECT COUNT(*) as total,
    COUNT(*) FILTER (
        WHERE status = 'pending'
    ) as pending,
    COUNT(*) FILTER (
        WHERE status = 'in_progress'
    ) as in_progress,
    COUNT(*) FILTER (
        WHERE status = 'completed'
    ) as completed,
    COUNT(*) FILTER (
        WHERE due_date < NOW()
            AND status != 'completed'
    ) as overdue
FROM todos
WHERE user_id = $1;
-- name: HardDeleteTodo :exec
DELETE FROM todos
WHERE id = $1
    AND user_id = $2;
-- name: HardDeleteTodos :exec
DELETE FROM todos
WHERE id = ANY($1::uuid [])
    AND user_id = $2;