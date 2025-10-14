# 🌍 Real-World Examples: Advanced sqlc + pgx Patterns

Practical examples for common scenarios you'll encounter.

## Example 1: User Registration with Transaction

**Scenario**: Create user, send verification email, log activity - all or nothing.

```go
// service/user_service.go
func (s *UserService) RegisterUser(ctx context.Context, input CreateUserInput) (*UserResponse, error) {
    var user db.User

    // Use transaction to ensure atomicity
    err := database.WithTransaction(ctx, s.pool.Pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
        txQueries := s.queries.WithTx(tx)

        // 1. Create user
        hashedPwd, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
        if err != nil {
            return err
        }

        user, err = txQueries.CreateUser(ctx, db.CreateUserParams{
            ID:           uuid.New(),
            Email:        input.Email,
            Username:     input.Username,
            PasswordHash: string(hashedPwd),
            Role:         db.UserRoleUser,
            Metadata:     []byte("{}"),
        })
        if err != nil {
            return fmt.Errorf("failed to create user: %w", err)
        }

        // 2. Create verification token
        token := uuid.New().String()
        // Store token logic here...

        // 3. Log registration activity
        err = txQueries.CreateActivityLog(ctx, db.CreateActivityLogParams{
            UserID: pgtype.UUID{Bytes: user.ID, Valid: true},
            Action: "user_registered",
            ResourceType: "user",
            ResourceID: user.ID,
            Metadata: []byte(`{"verification_sent": true}`),
        })
        if err != nil {
            return fmt.Errorf("failed to log activity: %w", err)
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    // Send verification email (outside transaction - can retry)
    // s.emailService.SendVerification(user.Email, token)

    return s.toUserResponse(user), nil
}
```

## Example 2: Complex Todo Filtering

**Scenario**: Filter todos by multiple criteria (status, priority, tags, due date).

```sql
-- queries/todos.sql
-- name: FilterTodos :many
SELECT * FROM todos
WHERE user_id = $1
  AND ($2::todo_status IS NULL OR status = $2)
  AND ($3::todo_priority IS NULL OR priority = $3)
  AND ($4::text[] IS NULL OR tags && $4)
  AND ($5::timestamptz IS NULL OR due_date <= $5)
  AND deleted_at IS NULL
ORDER BY
    CASE priority
        WHEN 'urgent' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END,
    due_date ASC NULLS LAST,
    created_at DESC
LIMIT $6 OFFSET $7;
```

```go
// service/todo_service.go
type TodoFilter struct {
    UserID   uuid.UUID
    Status   *string
    Priority *string
    Tags     []string
    DueBefore *time.Time
    Limit    int32
    Offset   int32
}

func (s *TodoService) FilterTodos(ctx context.Context, filter TodoFilter) ([]*TodoResponse, error) {
    // Build params with NULL for unset filters
    var status pgtype.Text
    if filter.Status != nil {
        status = pgtype.Text{String: *filter.Status, Valid: true}
    }

    var priority pgtype.Text
    if filter.Priority != nil {
        priority = pgtype.Text{String: *filter.Priority, Valid: true}
    }

    var tags []string
    if len(filter.Tags) > 0 {
        tags = filter.Tags
    }

    var dueBefore pgtype.Timestamptz
    if filter.DueBefore != nil {
        dueBefore = pgtype.Timestamptz{Time: *filter.DueBefore, Valid: true}
    }

    todos, err := s.queries.FilterTodos(ctx, db.FilterTodosParams{
        UserID:  filter.UserID,
        Column2: status,
        Column3: priority,
        Column4: tags,
        Column5: dueBefore,
        Limit:   filter.Limit,
        Offset:  filter.Offset,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to filter todos: %w", err)
    }

    return s.toTodoResponses(todos), nil
}
```

## Example 3: Optimistic Locking with Version

**Scenario**: Prevent concurrent updates using version field.

```sql
-- Add version column to todos table
ALTER TABLE todos ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- name: UpdateTodoWithVersion :one
UPDATE todos
SET
    title = $2,
    description = $3,
    status = $4,
    version = version + 1,
    updated_at = NOW()
WHERE id = $1
  AND version = $5  -- Only update if version matches
  AND deleted_at IS NULL
RETURNING *;
```

```go
func (s *TodoService) UpdateTodo(ctx context.Context, id uuid.UUID, version int32, input UpdateTodoInput) (*TodoResponse, error) {
    todo, err := s.queries.UpdateTodoWithVersion(ctx, db.UpdateTodoWithVersionParams{
        ID:          id,
        Title:       input.Title,
        Description: pgtype.Text{String: input.Description, Valid: true},
        Status:      db.TodoStatus(input.Status),
        Version:     version,
    })
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("todo not found or version mismatch (concurrent update)")
        }
        return nil, err
    }

    return s.toTodoResponse(todo), nil
}
```

## Example 4: Pagination with Cursor

**Scenario**: Efficient pagination using cursor-based approach.

```sql
-- name: ListTodosCursor :many
SELECT * FROM todos
WHERE user_id = $1
  AND ($2::timestamptz IS NULL OR created_at < $2)  -- Cursor
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3;
```

```go
type PaginatedResponse struct {
    Data       []*TodoResponse `json:"data"`
    NextCursor *time.Time      `json:"next_cursor,omitempty"`
    HasMore    bool            `json:"has_more"`
}

func (s *TodoService) ListTodosPaginated(ctx context.Context, userID uuid.UUID, cursor *time.Time, limit int32) (*PaginatedResponse, error) {
    var cursorParam pgtype.Timestamptz
    if cursor != nil {
        cursorParam = pgtype.Timestamptz{Time: *cursor, Valid: true}
    }

    // Fetch one extra to check if there are more
    todos, err := s.queries.ListTodosCursor(ctx, db.ListTodosCursorParams{
        UserID:  userID,
        Column2: cursorParam,
        Limit:   limit + 1,
    })
    if err != nil {
        return nil, err
    }

    hasMore := len(todos) > int(limit)
    if hasMore {
        todos = todos[:limit]
    }

    var nextCursor *time.Time
    if hasMore && len(todos) > 0 {
        lastTodo := todos[len(todos)-1]
        nextCursor = &lastTodo.CreatedAt
    }

    return &PaginatedResponse{
        Data:       s.toTodoResponses(todos),
        NextCursor: nextCursor,
        HasMore:    hasMore,
    }, nil
}
```

## Example 5: Aggregate User Dashboard

**Scenario**: Get comprehensive user statistics in one query.

```sql
-- name: GetUserDashboard :one
SELECT
    u.id,
    u.email,
    u.username,
    COUNT(DISTINCT t.id) as total_todos,
    COUNT(DISTINCT t.id) FILTER (WHERE t.status = 'completed') as completed_todos,
    COUNT(DISTINCT t.id) FILTER (WHERE t.due_date < NOW() AND t.status != 'completed') as overdue_todos,
    COUNT(DISTINCT t.id) FILTER (WHERE t.created_at > NOW() - INTERVAL '7 days') as todos_this_week,
    COALESCE(AVG(EXTRACT(EPOCH FROM (t.completed_at - t.created_at))), 0) as avg_completion_time_seconds
FROM users u
LEFT JOIN todos t ON u.id = t.user_id AND t.deleted_at IS NULL
WHERE u.id = $1
  AND u.deleted_at IS NULL
GROUP BY u.id, u.email, u.username;
```

```go
type UserDashboard struct {
    User                    *UserResponse `json:"user"`
    TotalTodos              int64         `json:"total_todos"`
    CompletedTodos          int64         `json:"completed_todos"`
    OverdueTodos            int64         `json:"overdue_todos"`
    TodosThisWeek           int64         `json:"todos_this_week"`
    AvgCompletionTimeSeconds float64      `json:"avg_completion_time_seconds"`
}

func (s *UserService) GetDashboard(ctx context.Context, userID uuid.UUID) (*UserDashboard, error) {
    stats, err := s.queries.GetUserDashboard(ctx, userID)
    if err != nil {
        return nil, err
    }

    return &UserDashboard{
        User: &UserResponse{
            ID:       stats.ID,
            Email:    stats.Email,
            Username: stats.Username,
        },
        TotalTodos:               stats.TotalTodos,
        CompletedTodos:           stats.CompletedTodos,
        OverdueTodos:             stats.OverdueTodos,
        TodosThisWeek:            stats.TodosThisWeek,
        AvgCompletionTimeSeconds: stats.AvgCompletionTimeSeconds,
    }, nil
}
```

## Example 6: Bulk Operations with Transactions

**Scenario**: Complete multiple todos and update user stats atomically.

```go
func (s *TodoService) BulkCompleteTodosWithStats(ctx context.Context, todoIDs []uuid.UUID, userID uuid.UUID) error {
    return database.WithTransaction(ctx, s.pool.Pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
        txQueries := s.queries.WithTx(tx)

        // 1. Complete all todos
        err := txQueries.BulkUpdateTodoStatus(ctx, db.BulkUpdateTodoStatusParams{
            Column1: todoIDs,
            Column2: db.TodoStatusCompleted,
        })
        if err != nil {
            return fmt.Errorf("failed to update todos: %w", err)
        }

        // 2. Get updated stats
        stats, err := txQueries.GetTodoStats(ctx, userID)
        if err != nil {
            return fmt.Errorf("failed to get stats: %w", err)
        }

        // 3. Update user metadata with stats
        metadata := map[string]interface{}{
            "last_bulk_complete": time.Now(),
            "total_completed":    stats.Completed,
        }
        metadataJSON, _ := json.Marshal(metadata)

        _, err = txQueries.UpdateUser(ctx, db.UpdateUserParams{
            ID: userID,
            Metadata: pgtype.JSONB{
                Bytes: metadataJSON,
                Valid: true,
            },
        })
        if err != nil {
            return fmt.Errorf("failed to update user: %w", err)
        }

        // 4. Log bulk action
        err = txQueries.CreateActivityLog(ctx, db.CreateActivityLogParams{
            UserID: pgtype.UUID{Bytes: userID, Valid: true},
            Action: "bulk_complete",
            ResourceType: "todo",
            ResourceID: todoIDs[0], // Primary resource
            Metadata: []byte(fmt.Sprintf(`{"count": %d}`, len(todoIDs))),
        })

        return err
    })
}
```

## Example 7: Search with Ranking

**Scenario**: Search todos with relevance ranking.

```sql
-- name: SearchTodosRanked :many
SELECT
    *,
    ts_rank(
        to_tsvector('english', title || ' ' || COALESCE(description, '')),
        plainto_tsquery('english', $2)
    ) as rank
FROM todos
WHERE user_id = $1
  AND to_tsvector('english', title || ' ' || COALESCE(description, ''))
      @@ plainto_tsquery('english', $2)
  AND deleted_at IS NULL
ORDER BY rank DESC, created_at DESC
LIMIT $3;
```

```go
func (s *TodoService) SearchTodosRanked(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]*TodoResponse, error) {
    todos, err := s.queries.SearchTodosRanked(ctx, db.SearchTodosRankedParams{
        UserID: userID,
        Column2: query,
        Limit: limit,
    })
    if err != nil {
        return nil, err
    }

    return s.toTodoResponses(todos), nil
}
```

## Example 8: Handling Race Conditions

**Scenario**: Increment counter safely with row locking.

```sql
-- name: IncrementTodoViewCount :one
SELECT view_count FROM todos
WHERE id = $1
FOR UPDATE;

-- name: UpdateTodoViewCount :exec
UPDATE todos
SET view_count = $2
WHERE id = $1;
```

```go
func (s *TodoService) IncrementViewCount(ctx context.Context, todoID uuid.UUID) error {
    return database.WithTransaction(ctx, s.pool.Pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
        txQueries := s.queries.WithTx(tx)

        // Lock row and get current count
        currentCount, err := txQueries.IncrementTodoViewCount(ctx, todoID)
        if err != nil {
            return err
        }

        // Update with incremented value
        return txQueries.UpdateTodoViewCount(ctx, db.UpdateTodoViewCountParams{
            ID:        todoID,
            ViewCount: currentCount + 1,
        })
    })
}
```

## Example 9: Soft Delete with Cascade

**Scenario**: Soft delete user and all related data.

```go
func (s *UserService) DeleteUserWithData(ctx context.Context, userID uuid.UUID) error {
    return database.WithTransaction(ctx, s.pool.Pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
        txQueries := s.queries.WithTx(tx)

        // 1. Soft delete all user's todos
        err := txQueries.BulkDeleteTodos(ctx, db.BulkDeleteTodosParams{
            Column1: []uuid.UUID{},  // Will be fetched
            Column2: userID,
        })
        if err != nil {
            return err
        }

        // 2. Soft delete user
        err = txQueries.SoftDeleteUser(ctx, userID)
        if err != nil {
            return err
        }

        // 3. Log deletion
        return txQueries.CreateActivityLog(ctx, db.CreateActivityLogParams{
            UserID: pgtype.UUID{Bytes: userID, Valid: true},
            Action: "user_deleted",
            ResourceType: "user",
            ResourceID: userID,
            Metadata: []byte(`{"with_data": true}`),
        })
    })
}
```

## Example 10: Scheduled Jobs with Advisory Locks

**Scenario**: Process overdue todos without conflicts between multiple workers.

```sql
-- name: LockAndGetOverdueTodos :many
SELECT * FROM todos
WHERE status != 'completed'
  AND due_date < NOW()
  AND deleted_at IS NULL
  AND pg_try_advisory_xact_lock(('x' || substr(md5(id::text), 1, 16))::bit(64)::bigint)
LIMIT $1
FOR UPDATE SKIP LOCKED;
```

```go
func (s *TodoService) ProcessOverdueTodos(ctx context.Context, limit int32) error {
    return database.WithTransaction(ctx, s.pool.Pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
        txQueries := s.queries.WithTx(tx)

        // Get and lock overdue todos (advisory lock prevents other workers)
        todos, err := txQueries.LockAndGetOverdueTodos(ctx, limit)
        if err != nil {
            return err
        }

        for _, todo := range todos {
            // Send notification
            // s.notificationService.SendOverdueNotification(todo)

            // Update metadata
            metadata := map[string]interface{}{
                "overdue_notification_sent": time.Now(),
            }
            metadataJSON, _ := json.Marshal(metadata)

            _, err = txQueries.UpdateTodo(ctx, db.UpdateTodoParams{
                ID: todo.ID,
                Metadata: pgtype.JSONB{
                    Bytes: metadataJSON,
                    Valid: true,
                },
            })
            if err != nil {
                return err
            }
        }

        return nil
    })
}
```

## Example 11: Audit Trail with Change Tracking

**Scenario**: Track what changed when updating a record.

```go
func (s *TodoService) UpdateTodoWithAudit(ctx context.Context, todoID uuid.UUID, userID uuid.UUID, input UpdateTodoInput) (*TodoResponse, error) {
    var updatedTodo db.Todo

    err := database.WithTransaction(ctx, s.pool.Pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
        txQueries := s.queries.WithTx(tx)

        // Get current state
        oldTodo, err := txQueries.GetTodoByID(ctx, todoID)
        if err != nil {
            return err
        }

        // Update todo
        updatedTodo, err = txQueries.UpdateTodo(ctx, db.UpdateTodoParams{
            ID:    todoID,
            Title: pgtype.Text{String: input.Title, Valid: true},
            Status: pgtype.Text{String: input.Status, Valid: true},
        })
        if err != nil {
            return err
        }

        // Track changes
        changes := map[string]interface{}{}
        if oldTodo.Title != updatedTodo.Title {
            changes["title"] = map[string]string{
                "old": oldTodo.Title,
                "new": updatedTodo.Title,
            }
        }
        if oldTodo.Status != updatedTodo.Status {
            changes["status"] = map[string]string{
                "old": string(oldTodo.Status),
                "new": string(updatedTodo.Status),
            }
        }

        changesJSON, _ := json.Marshal(changes)

        // Log changes
        return txQueries.CreateActivityLog(ctx, db.CreateActivityLogParams{
            UserID: pgtype.UUID{Bytes: userID, Valid: true},
            Action: "todo_updated",
            ResourceType: "todo",
            ResourceID: todoID,
            Metadata: changesJSON,
        })
    })

    if err != nil {
        return nil, err
    }

    return s.toTodoResponse(updatedTodo), nil
}
```

## Example 12: Rate Limiting with PostgreSQL

**Scenario**: Rate limit API calls using database.

```sql
CREATE TABLE rate_limits (
    user_id UUID NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    request_count INTEGER NOT NULL DEFAULT 1,
    window_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, endpoint)
);

-- name: CheckRateLimit :one
INSERT INTO rate_limits (user_id, endpoint, request_count, window_start)
VALUES ($1, $2, 1, NOW())
ON CONFLICT (user_id, endpoint) DO UPDATE
SET
    request_count = CASE
        WHEN rate_limits.window_start < NOW() - INTERVAL '1 minute'
        THEN 1
        ELSE rate_limits.request_count + 1
    END,
    window_start = CASE
        WHEN rate_limits.window_start < NOW() - INTERVAL '1 minute'
        THEN NOW()
        ELSE rate_limits.window_start
    END
RETURNING request_count, window_start;
```

```go
func (s *UserService) CheckRateLimit(ctx context.Context, userID uuid.UUID, endpoint string, maxRequests int32) error {
    limit, err := s.queries.CheckRateLimit(ctx, db.CheckRateLimitParams{
        UserID:   userID,
        Endpoint: endpoint,
    })
    if err != nil {
        return err
    }

    if limit.RequestCount > maxRequests {
        return fmt.Errorf("rate limit exceeded: %d requests in window", limit.RequestCount)
    }

    return nil
}
```

---

## 🎓 Key Takeaways

1. **Transactions** are your friend for multi-step operations
2. **Row locking** prevents race conditions
3. **Cursors** are more efficient than offset pagination
4. **Aggregations** reduce round trips
5. **Advisory locks** coordinate distributed workers
6. **Audit trails** provide accountability
7. **Version fields** enable optimistic locking
8. **Soft deletes** preserve data integrity

Use these patterns as templates for your own features! 🚀
