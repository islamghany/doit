# 📝 sqlc + pgx Cheat Sheet

Quick reference for common patterns and operations.

## 🔧 Common sqlc Query Patterns

### Basic CRUD

```sql
-- name: CreateUser :one
INSERT INTO users (id, email, username, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUser :exec
UPDATE users SET email = $2 WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
```

### Return Types

- `:one` - Returns single row (error if not found)
- `:many` - Returns slice of rows
- `:exec` - Returns nothing (affected rows via error)
- `:execrows` - Returns number of affected rows
- `:execresult` - Returns pgconn.CommandTag

### Optional Parameters

```sql
-- name: UpdateUserOptional :one
UPDATE users
SET
    email = COALESCE(sqlc.narg('email'), email),
    username = COALESCE(sqlc.narg('username'), username),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;
```

### Pagination

```sql
-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
```

### Aggregations

```sql
-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL;

-- name: GetUserStats :one
SELECT
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE is_active = true) as active,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') as new_this_week
FROM users;
```

### Joins

```sql
-- name: GetUserWithProfile :one
SELECT
    u.*,
    p.bio,
    p.avatar_url
FROM users u
LEFT JOIN profiles p ON u.id = p.user_id
WHERE u.id = $1;

-- name: ListTodosWithUser :many
SELECT
    t.*,
    u.username,
    u.email
FROM todos t
INNER JOIN users u ON t.user_id = u.id
WHERE t.deleted_at IS NULL
ORDER BY t.created_at DESC;
```

### Array Operations

```sql
-- name: GetTodosByTags :many
SELECT * FROM todos
WHERE tags && $1::text[]  -- Overlaps
  AND deleted_at IS NULL;

-- name: GetTodosWithAllTags :many
SELECT * FROM todos
WHERE tags @> $1::text[]  -- Contains all
  AND deleted_at IS NULL;
```

### JSONB Operations

```sql
-- name: GetUsersByMetadataKey :many
SELECT * FROM users
WHERE metadata ? $1  -- Has key
  AND deleted_at IS NULL;

-- name: GetUsersByMetadataValue :many
SELECT * FROM users
WHERE metadata @> $1::jsonb  -- Contains
  AND deleted_at IS NULL;

-- name: UpdateUserMetadata :exec
UPDATE users
SET metadata = metadata || $2::jsonb
WHERE id = $1;
```

### Full-Text Search

```sql
-- name: SearchTodosByTitle :many
SELECT * FROM todos
WHERE title ILIKE '%' || $1 || '%'
  AND deleted_at IS NULL
LIMIT $2;

-- Using pg_trgm for similarity
-- name: SearchSimilarTodos :many
SELECT *, similarity(title, $1) as sim
FROM todos
WHERE title % $1
ORDER BY sim DESC
LIMIT $2;
```

### Row Locking

```sql
-- name: GetTodoForUpdate :one
SELECT * FROM todos
WHERE id = $1
FOR UPDATE;  -- Exclusive lock

-- name: GetTodoForShare :one
SELECT * FROM todos
WHERE id = $1
FOR SHARE;  -- Shared lock
```

## 🔄 Transaction Patterns

### Basic Transaction

```go
err := database.WithTransaction(ctx, pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
    queries := db.New(tx)

    user, err := queries.CreateUser(ctx, params)
    if err != nil {
        return err  // Automatic rollback
    }

    err = queries.CreateProfile(ctx, profileParams)
    return err  // Automatic commit if nil
})
```

### Serializable Transaction

```go
err := database.WithSerializableTransaction(ctx, pool, func(tx pgx.Tx) error {
    queries := db.New(tx)
    // High isolation operations
    return nil
})
```

### Read-Only Transaction

```go
err := database.WithReadOnlyTransaction(ctx, pool, func(tx pgx.Tx) error {
    queries := db.New(tx)
    // Read-only operations
    return nil
})
```

### Custom Transaction Options

```go
opts := database.TxOptions{
    IsoLevel:   pgx.RepeatableRead,
    AccessMode: pgx.ReadWrite,
}

err := database.WithTransaction(ctx, pool, opts, func(tx pgx.Tx) error {
    // Your operations
    return nil
})
```

## 📦 Batch Operations

### Batch Queries

```go
batch := &pgx.Batch{}
batch.Queue("INSERT INTO users (id, email) VALUES ($1, $2)", uuid.New(), "user1@example.com")
batch.Queue("INSERT INTO users (id, email) VALUES ($1, $2)", uuid.New(), "user2@example.com")
batch.Queue("INSERT INTO users (id, email) VALUES ($1, $2)", uuid.New(), "user3@example.com")

results := pool.SendBatch(ctx, batch)
defer results.Close()

// Process results
for i := 0; i < 3; i++ {
    _, err := results.Exec()
    if err != nil {
        return err
    }
}
```

### Bulk Insert with CopyFrom

```go
rows := [][]interface{}{
    {uuid.New(), "user1@example.com", "user1", "hash1"},
    {uuid.New(), "user2@example.com", "user2", "hash2"},
    {uuid.New(), "user3@example.com", "user3", "hash3"},
}

count, err := pool.CopyFrom(
    ctx,
    pgx.Identifier{"users"},
    []string{"id", "email", "username", "password_hash"},
    pgx.CopyFromRows(rows),
)
```

## 🎯 NULL Handling

### Reading NULL Values

```go
// pgtype automatically generated by sqlc
if user.LastLoginAt.Valid {
    fmt.Println("Last login:", user.LastLoginAt.Time)
} else {
    fmt.Println("Never logged in")
}
```

### Writing NULL Values

```go
// Nil value
params := db.CreateTodoParams{
    // ...
    DueDate: pgtype.Timestamptz{}, // NULL
}

// With value
params := db.CreateTodoParams{
    // ...
    DueDate: pgtype.Timestamptz{
        Time:  time.Now().Add(24 * time.Hour),
        Valid: true,
    },
}

// Or use pointer for optional fields
var dueDate *time.Time
if input.DueDate != nil {
    dueDate = input.DueDate
}
```

## 🔍 Error Handling

### Check for No Rows

```go
user, err := queries.GetUserByID(ctx, userID)
if err != nil {
    if err == pgx.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    return nil, fmt.Errorf("database error: %w", err)
}
```

### PostgreSQL Error Codes

```go
import "github.com/jackc/pgx/v5/pgconn"

if pgErr, ok := err.(*pgconn.PgError); ok {
    switch pgErr.Code {
    case "23505":  // unique_violation
        return fmt.Errorf("duplicate entry")
    case "23503":  // foreign_key_violation
        return fmt.Errorf("referenced resource not found")
    case "23502":  // not_null_violation
        return fmt.Errorf("required field missing")
    case "22001":  // string_data_right_truncation
        return fmt.Errorf("value too long")
    default:
        return fmt.Errorf("database error: %s", pgErr.Message)
    }
}
```

### Common Error Codes

- `23505` - Unique violation
- `23503` - Foreign key violation
- `23502` - NOT NULL violation
- `22001` - String too long
- `42P01` - Table does not exist
- `42703` - Column does not exist

## 🧪 Testing Patterns

### Service Test with Real DB

```go
func TestUserService_CreateUser(t *testing.T) {
    ctx := context.Background()
    pool := setupTestDB(t)
    defer pool.Close()

    service := service.NewUserService(pool)

    input := service.CreateUserInput{
        Email:    "test@example.com",
        Username: "testuser",
        Password: "password123",
        Role:     "user",
    }

    user, err := service.CreateUser(ctx, input)
    require.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}

func setupTestDB(t *testing.T) *database.Pool {
    cfg := database.Config{
        Host:     "localhost",
        Port:     5432,
        Database: "doit_test",
        User:     "doit",
        Password: "doit123",
        MaxConns: 5,
        MinConns: 1,
    }

    pool, err := database.New(context.Background(), cfg)
    require.NoError(t, err)

    // Run migrations
    // ...

    t.Cleanup(func() {
        // Clean up test data
        pool.Close()
    })

    return pool
}
```

## 💡 Performance Tips

### Connection Pool Settings

```go
// Production settings
poolConfig.MaxConns = 25                          // Don't set too high
poolConfig.MinConns = 5                           // Keep warm connections
poolConfig.MaxConnLifetime = 1 * time.Hour        // Recycle connections
poolConfig.MaxConnIdleTime = 30 * time.Minute     // Close idle
poolConfig.HealthCheckPeriod = 1 * time.Minute    // Regular checks
```

### Query Optimization

```sql
-- Use LIMIT
SELECT * FROM users LIMIT 100;

-- Use indexes
CREATE INDEX idx_users_email ON users(email);

-- Partial indexes
CREATE INDEX idx_active_users ON users(status) WHERE deleted_at IS NULL;

-- Covering indexes (index-only scans)
CREATE INDEX idx_user_details ON users(id) INCLUDE (email, username);

-- GIN indexes for arrays/JSONB
CREATE INDEX idx_todos_tags ON todos USING gin(tags);
CREATE INDEX idx_users_metadata ON users USING gin(metadata);
```

### Prepared Statements

```go
// sqlc automatically uses prepared statements when:
// emit_prepared_queries: true in sqlc.yaml

// First call: prepares statement
user, err := queries.GetUserByID(ctx, userID)

// Subsequent calls: reuses prepared statement (faster)
user2, err := queries.GetUserByID(ctx, anotherID)
```

### Use CopyFrom for Bulk Inserts

```go
// CopyFrom is 10-100x faster than individual inserts
count, err := pool.CopyFrom(ctx, ...)
```

## 🛠️ Utility Functions

### Check Connection

```go
func (p *Pool) Ping(ctx context.Context) error {
    return p.Pool.Ping(ctx)
}
```

### Get Stats

```go
stats := pool.Stats()
fmt.Printf("Total connections: %d\n", stats.TotalConns())
fmt.Printf("Idle connections: %d\n", stats.IdleConns())
fmt.Printf("Acquired connections: %d\n", stats.AcquiredConns())
```

### Execute Raw Query

```go
// For queries not in sqlc
rows, err := pool.Query(ctx, "SELECT * FROM users WHERE created_at > $1", time.Now().Add(-24*time.Hour))
defer rows.Close()

for rows.Next() {
    var user User
    err := rows.Scan(&user.ID, &user.Email, ...)
    // Process user
}
```

## 🎓 Best Practices

### ✅ DO

- Keep queries in SQL files
- Use transactions for multi-step operations
- Handle NULL values properly
- Use prepared statements (automatic with sqlc)
- Use connection pooling
- Add indexes for frequently queried columns
- Use partial indexes for filtered queries
- Handle PostgreSQL errors explicitly
- Test with real database
- Use CopyFrom for bulk inserts

### ❌ DON'T

- Don't edit generated code
- Don't use ORM query builders alongside sqlc
- Don't set MaxConns too high
- Don't forget to close connections/rows
- Don't ignore transaction errors
- Don't use SELECT \* in production (specify columns)
- Don't create indexes on every column
- Don't ignore N+1 query problems
- Don't test only with mocks

## 📚 Common Recipes

### Upsert

```sql
-- name: UpsertUser :one
INSERT INTO users (id, email, username, password_hash)
VALUES ($1, $2, $3, $4)
ON CONFLICT (email) DO UPDATE
SET username = EXCLUDED.username,
    updated_at = NOW()
RETURNING *;
```

### Soft Delete

```sql
-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW() WHERE id = $1;

-- Always filter in queries
-- name: GetActiveUsers :many
SELECT * FROM users WHERE deleted_at IS NULL;
```

### Audit Trail

```sql
-- name: CreateWithAudit :one
WITH inserted AS (
    INSERT INTO todos (id, user_id, title)
    VALUES ($1, $2, $3)
    RETURNING *
)
INSERT INTO audit_logs (resource_type, resource_id, action, user_id)
SELECT 'todo', id, 'created', user_id
FROM inserted
RETURNING *;
```

### Conditional Update

```sql
-- name: IncrementViewCount :exec
UPDATE articles
SET view_count = view_count + 1
WHERE id = $1
  AND view_count < 1000;  -- Only if under limit
```

## 🔗 Quick Links

- [sqlc Documentation](https://docs.sqlc.dev)
- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [PostgreSQL Error Codes](https://www.postgresql.org/docs/current/errcodes-appendix.html)
- [PostgreSQL Performance](https://wiki.postgresql.org/wiki/Performance_Optimization)

---

**Keep this cheat sheet handy for quick reference! 📌**
