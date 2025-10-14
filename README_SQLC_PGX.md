# 🚀 Unlocking Full Potential of sqlc + pgx

This guide demonstrates advanced features and best practices for using sqlc with pgx in a **Fat Service Pattern** architecture.

## 📋 Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Advanced Features](#advanced-features)
3. [Performance Optimizations](#performance-optimizations)
4. [Best Practices](#best-practices)
5. [Setup Instructions](#setup-instructions)

---

## 🏗️ Architecture Overview

### Fat Service Pattern Structure

```
internal/
├── service/           # Fat services (all business logic here)
│   ├── user_service.go
│   ├── todo_service.go
│   └── ...
├── data/
│   ├── db/           # Generated sqlc code (DO NOT EDIT)
│   ├── queries/      # SQL queries (you write these)
│   └── migrations/   # Database migrations
pkg/
├── database/         # Database connection & utilities
```

**Key Principle**: Services contain ALL business logic. No thin repository layer.

---

## 🔥 Advanced Features

### 1. **Transaction Management**

```go
// Automatic rollback on error, commit on success
err := database.WithTransaction(ctx, pool, opts, func(tx pgx.Tx) error {
    txQueries := queries.WithTx(tx)

    // Multiple operations in transaction
    user, err := txQueries.CreateUser(ctx, params)
    if err != nil {
        return err // Auto-rollback
    }

    err = txQueries.CreateActivityLog(ctx, logParams)
    return err // Auto-commit if nil
})
```

**Advanced Transaction Types**:

- `WithSerializableTransaction` - Highest isolation level
- `WithReadOnlyTransaction` - Optimized for read-only queries
- Custom isolation levels and access modes

### 2. **Row Locking (Optimistic & Pessimistic)**

```sql
-- queries/todos.sql
-- name: GetTodoByIDForUpdate :one
SELECT * FROM todos
WHERE id = $1 AND deleted_at IS NULL
FOR UPDATE;  -- Locks the row until transaction completes
```

Use cases:

- Prevent race conditions in updates
- Ensure consistency in multi-step operations
- Implement optimistic locking

### 3. **Batch Operations**

```go
// Batch inserts using CopyFrom (10-100x faster)
copyCount, err := pool.CopyFrom(
    ctx,
    pgx.Identifier{"todos"},
    []string{"id", "user_id", "title", ...},
    pgx.CopyFromRows(rows),
)
```

**When to use**:

- Bulk imports
- Data migrations
- Seeding test data

### 4. **PostgreSQL-Specific Types**

#### JSONB Operations

```sql
-- Merge JSONB
UPDATE users
SET metadata = metadata || $2::jsonb
WHERE id = $1;
```

#### Array Operations

```sql
-- Array contains (tags)
SELECT * FROM todos
WHERE tags && $1::text[]  -- Overlaps with input array
```

#### Enums

```sql
CREATE TYPE todo_status AS ENUM ('pending', 'in_progress', 'completed');
```

### 5. **Advanced Queries**

#### Aggregations

```sql
-- name: GetTodoStats :one
SELECT
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE status = 'completed') as completed,
    COUNT(*) FILTER (WHERE due_date < NOW()) as overdue
FROM todos
WHERE user_id = $1;
```

#### Joins with Aggregations

```sql
-- name: ListTodosWithAttachmentCount :many
SELECT
    t.*,
    COUNT(a.id) as attachment_count
FROM todos t
LEFT JOIN todo_attachments a ON t.id = a.todo_id
GROUP BY t.id;
```

#### Full-Text Search

```sql
-- Using pg_trgm extension
SELECT * FROM todos
WHERE title ILIKE '%' || $1 || '%'
```

### 6. **Prepared Statements**

sqlc automatically generates prepared statements when `emit_prepared_queries: true`:

```go
// First call: prepares statement
// Subsequent calls: reuse prepared statement (faster)
todo, err := queries.GetTodoByID(ctx, todoID)
```

### 7. **NULL Handling**

```go
// Optional fields with pgtype
var dueDate pgtype.Timestamptz
if input.DueDate != nil {
    dueDate = pgtype.Timestamptz{
        Time:  *input.DueDate,
        Valid: true,
    }
}
```

### 8. **Named Parameters**

```sql
-- name: UpdateTodo :one
UPDATE todos
SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    priority = COALESCE(sqlc.narg('priority'), priority)
WHERE id = sqlc.arg('id')
RETURNING *;
```

Benefits:

- Optional updates (only update provided fields)
- Type-safe parameter names
- Better code readability

---

## ⚡ Performance Optimizations

### 1. **Connection Pooling**

```go
poolConfig.MaxConns = 25                          // Max connections
poolConfig.MinConns = 5                           // Min idle connections
poolConfig.MaxConnLifetime = 1 * time.Hour        // Recycle old connections
poolConfig.MaxConnIdleTime = 30 * time.Minute     // Close idle connections
poolConfig.HealthCheckPeriod = 1 * time.Minute    // Regular health checks
```

### 2. **Query Execution Modes**

```go
// Cache prepared statements (default in our setup)
poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
```

Modes:

- `QueryExecModeCacheStatement` - Best for repeated queries
- `QueryExecModeSimpleProtocol` - Simple queries
- `QueryExecModeExec` - One-time queries

### 3. **Batch Queries**

```go
batch := &pgx.Batch{}
batch.Queue("INSERT INTO todos ...", args1...)
batch.Queue("INSERT INTO todos ...", args2...)

results := pool.SendBatch(ctx, batch)
defer results.Close()
```

### 4. **Index Strategies**

```sql
-- Partial indexes (smaller, faster)
CREATE INDEX idx_active_todos ON todos(user_id, status)
WHERE deleted_at IS NULL;

-- Covering indexes (index-only scans)
CREATE INDEX idx_todo_details ON todos(user_id)
INCLUDE (title, status, priority);

-- GIN indexes for JSONB/arrays
CREATE INDEX idx_todos_metadata ON todos USING gin(metadata);
CREATE INDEX idx_todos_tags ON todos USING gin(tags);
```

### 5. **Query Optimization Tips**

- Use `LIMIT` and `OFFSET` for pagination
- Add `WHERE deleted_at IS NULL` to partial indexes
- Use `EXPLAIN ANALYZE` to identify slow queries
- Avoid N+1 queries with JOINs
- Use CTEs for complex queries

---

## 🎯 Best Practices

### 1. **Service Layer Structure**

```go
type TodoService struct {
    pool    *database.Pool   // For transactions and custom queries
    queries *db.Queries      // Generated sqlc queries
}

// Always create queries from pool in constructor
func NewTodoService(pool *database.Pool) *TodoService {
    return &TodoService{
        pool:    pool,
        queries: db.New(pool),
    }
}
```

### 2. **Error Handling**

```go
if err == pgx.ErrNoRows {
    return nil, fmt.Errorf("resource not found")
}
if pgErr, ok := err.(*pgconn.PgError); ok {
    switch pgErr.Code {
    case "23505": // unique_violation
        return fmt.Errorf("duplicate entry")
    case "23503": // foreign_key_violation
        return fmt.Errorf("referenced resource not found")
    }
}
```

### 3. **Testing Services**

```go
// Use test database or mocks
func TestTodoService_CreateTodo(t *testing.T) {
    ctx := context.Background()
    pool := setupTestDB(t)
    defer pool.Close()

    service := NewTodoService(pool)

    todo, err := service.CreateTodo(ctx, input)
    assert.NoError(t, err)
    assert.NotEmpty(t, todo.ID)
}
```

### 4. **Migrations**

```bash
# Create migration
migrate create -ext sql -dir internal/data/migrations -seq add_users

# Run migrations
migrate -path internal/data/migrations \
        -database "postgresql://localhost/doit?sslmode=disable" \
        up
```

### 5. **Query Organization**

Group queries by domain:

- `queries/users.sql` - User operations
- `queries/todos.sql` - Todo operations
- `queries/activity_logs.sql` - Audit logs

---

## 🚀 Setup Instructions

### 1. Install Dependencies

```bash
# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install pgx
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
```

### 2. Update go.mod

```bash
go mod tidy
```

### 3. Start PostgreSQL

```bash
# Using Makefile
make dev-db

# Or manually
docker run --name doit-postgres \
  -e POSTGRES_USER=doit \
  -e POSTGRES_PASSWORD=doit123 \
  -e POSTGRES_DB=doit \
  -p 5432:5432 \
  -d postgres:16-alpine
```

### 4. Generate sqlc Code

```bash
make sqlc
# or
sqlc generate
```

### 5. Update Configuration

Uncomment database config in `internal/config/config.go`:

```go
type Config struct {
    Server   ServerConfig   `prefix:"SERVER_"`
    App      AppConfig      `prefix:"APP_"`
    Database DatabaseConfig `prefix:"DB_"`  // Uncomment this
}
```

### 6. Wire Up Services

```go
// cmd/doit/main.go
func main() {
    // ... existing code ...

    // Initialize database
    dbPool, err := database.New(ctx, database.Config{
        Host:            cfg.Database.Host,
        Port:            cfg.Database.Port,
        Database:        cfg.Database.Name,
        User:            cfg.Database.User,
        Password:        cfg.Database.Password,
        MaxConns:        cfg.Database.MaxOpenConns,
        MinConns:        5,
        MaxConnLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
        MaxConnIdleTime: 30 * time.Minute,
        DisableTLS:      cfg.Database.DisableTLS,
        LogLevel:        cfg.App.LogLevel,
    })
    if err != nil {
        log.Error(ctx, "Failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer dbPool.Close()

    // Initialize services
    userService := service.NewUserService(dbPool)
    todoService := service.NewTodoService(dbPool)

    // Pass to API
    if err := api.Run(ctx, log, cfg, userService, todoService); err != nil {
        log.Error(ctx, "Failed to start", "error", err)
        os.Exit(1)
    }
}
```

---

## 📚 Advanced Topics

### Listen/Notify (Real-time Events)

```go
conn, err := pool.Acquire(ctx)
defer conn.Release()

_, err = conn.Exec(ctx, "LISTEN todo_changes")

for {
    notification, err := conn.Conn().WaitForNotification(ctx)
    // Handle real-time notification
}
```

### Custom Type Mapping

```yaml
# sqlc.yaml
overrides:
  - db_type: "point"
    go_type: "github.com/paulmach/orb.Point"
```

### Dynamic Queries

For complex filtering, use query builders like `squirrel` or `goqu` alongside sqlc.

---

## 🎓 Learning Resources

- [sqlc Documentation](https://docs.sqlc.dev)
- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [PostgreSQL Performance Tips](https://wiki.postgresql.org/wiki/Performance_Optimization)

---

## 🤝 Contributing

When adding new features:

1. Write SQL in `internal/data/queries/`
2. Run `make sqlc` to generate code
3. Implement business logic in services
4. Write tests
5. Update this README if adding new patterns

---

**Happy Coding! 🎉**
