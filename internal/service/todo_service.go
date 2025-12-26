package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"doit/internal/data/db"
	"doit/internal/metrics"
	"doit/internal/model"
	"doit/internal/tracing"
	"doit/pkg/database"
	"doit/pkg/validator"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// TodoService errors
var (
	ErrTodoNotFound     = errors.New("todo not found")
	ErrTodoUnauthorized = errors.New("unauthorized: todo does not belong to user")
)

// TodoService handles all todo-related business logic
type TodoService struct {
	pool    *database.Pool
	querier db.Querier
}

func NewTodoService(pool *database.Pool) *TodoService {
	return &TodoService{
		pool:    pool,
		querier: db.New(pool),
	}
}

func NewTodoServiceWithQuerier(querier db.Querier) *TodoService {
	return &TodoService{
		querier: querier,
	}
}

// CreateTodo creates a new todo
func (s *TodoService) CreateTodo(ctx context.Context, input model.CreateTodoInput) (*model.Todo, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "create")
	defer span.End()

	// Validation span
	ctx, validateSpan := tracing.StartSpan(ctx, tracing.TracerService, "todo.validate")
	if err := s.validateCreateTodoInput(input); err != nil {
		tracing.RecordError(validateSpan, err)
		validateSpan.End()
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	validateSpan.End()

	// Prepare metadata
	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Set default priority if not provided
	priority := input.Priority
	if priority == "" {
		priority = model.TodoPriorityMedium
	}

	// Prepare optional fields
	var description *string
	if input.Description != "" {
		description = &input.Description
	}

	var dueDate pgtype.Timestamptz
	if input.DueDate != nil {
		dueDate = pgtype.Timestamptz{
			Time:  *input.DueDate,
			Valid: true,
		}
	}

	// Prepare tags
	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}
	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "insert", "todos")
	start := time.Now()
	todo, err := s.querier.CreateTodo(ctx, db.CreateTodoParams{
		ID:          uuid.New(),
		UserID:      input.UserID,
		Title:       input.Title,
		Description: description,
		Status:      db.TodoStatusPending,
		Priority:    db.TodoPriority(priority),
		Tags:        tags,
		Metadata:    metadataJSON,
		DueDate:     dueDate,
	})
	duration := time.Since(start).Seconds()
	metrics.RecordDatabaseQuery("create_todo", "todos", duration, err)
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		return nil, ErrTodoNotFound
	}
	tracing.SetAttributes(dbSpan, tracing.String("todo.id", todo.ID.String()))
	dbSpan.End()

	// Set result attributes on main span
	tracing.SetAttributes(span,
		tracing.String("todo.id", todo.ID.String()),
		tracing.String("user.id", input.UserID.String()),
	)

	// Record business metric
	metrics.RecordTodoOperation("create")
	return s.toTodoModel(todo), nil
}

// GetTodoByID retrieves a todo by ID
func (s *TodoService) GetTodoByID(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) (*model.Todo, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "get_by_id")
	defer span.End()
	tracing.SetAttributes(span,
		tracing.String("todo.id", todoID.String()),
		tracing.String("user.id", userID.String()),
	)

	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "select", "todos")
	start := time.Now()
	todo, err := s.querier.GetTodoByID(ctx, todoID)
	duration := time.Since(start).Seconds()
	metrics.RecordDatabaseQuery("get_todo_by_id", "todos", duration, err)
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		if err == pgx.ErrNoRows {
			return nil, ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}
	dbSpan.End()

	// Verify ownership
	if err = VerifyOwnership(todo.UserID, userID, "todo"); err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	// Record business metric
	metrics.RecordTodoOperation("read")
	return s.toTodoModel(todo), nil
}

// ListUserTodos retrieves paginated todos for a user
func (s *TodoService) ListUserTodos(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*model.Todo, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "list")
	defer span.End()
	tracing.SetAttributes(span,
		tracing.String("user.id", userID.String()),
		tracing.Int("pagination.limit", int(limit)),
		tracing.Int("pagination.offset", int(offset)),
	)

	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "select", "todos")
	start := time.Now()
	todos, err := s.querier.ListTodosByUser(ctx, db.ListTodosByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	duration := time.Since(start).Seconds()
	metrics.RecordDatabaseQuery("list_user_todos", "todos", duration, err)
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		return nil, fmt.Errorf("failed to list todos: %w", err)
	}
	tracing.SetAttributes(dbSpan, tracing.Int("db.rows_affected", len(todos)))
	dbSpan.End()

	// Set result count on main span
	tracing.SetAttributes(span, tracing.Int("result.count", len(todos)))

	// Record business metric
	metrics.RecordTodoOperation("list")
	return s.toTodoModels(todos), nil
}

// ListTodosByStatus retrieves todos filtered by status
func (s *TodoService) ListTodosByStatus(ctx context.Context, userID uuid.UUID, status model.TodoStatus, limit, offset int32) ([]*model.Todo, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "list_by_status")
	defer span.End()
	tracing.SetAttributes(span,
		tracing.String("user.id", userID.String()),
		tracing.String("filter.status", string(status)),
		tracing.Int("pagination.limit", int(limit)),
		tracing.Int("pagination.offset", int(offset)),
	)

	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "select", "todos")
	todos, err := s.querier.ListTodosByUserAndStatus(ctx, db.ListTodosByUserAndStatusParams{
		UserID: userID,
		Status: db.TodoStatus(status),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		return nil, fmt.Errorf("failed to list todos by status: %w", err)
	}
	tracing.SetAttributes(dbSpan, tracing.Int("db.rows_affected", len(todos)))
	dbSpan.End()

	tracing.SetAttributes(span, tracing.Int("result.count", len(todos)))
	return s.toTodoModels(todos), nil
}

// UpdateTodo updates a todo
func (s *TodoService) UpdateTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID, input model.UpdateTodoInput) (*model.Todo, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "update")
	defer span.End()
	tracing.SetAttributes(span,
		tracing.String("todo.id", todoID.String()),
		tracing.String("user.id", userID.String()),
	)

	// First, verify ownership - Database span for get
	ctx, getSpan := tracing.StartDBSpan(ctx, "select", "todos")
	existingTodo, err := s.querier.GetTodoByID(ctx, todoID)
	if err != nil {
		tracing.RecordError(getSpan, err)
		getSpan.End()
		if err == pgx.ErrNoRows {
			return nil, ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}
	getSpan.End()

	if err = VerifyOwnership(existingTodo.UserID, userID, "todo"); err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	// Build update params
	params := db.UpdateTodoParams{
		ID: todoID,
	}

	if input.Title != nil {
		params.Title = input.Title
	}

	if input.Description != nil {
		params.Description = input.Description
	}

	if input.Status != nil {
		status := db.TodoStatus(*input.Status)
		params.Status = db.NullTodoStatus{
			TodoStatus: status,
			Valid:      true,
		}
	}

	if input.Priority != nil {
		priority := db.TodoPriority(*input.Priority)
		params.Priority = db.NullTodoPriority{
			TodoPriority: priority,
			Valid:        true,
		}
	}

	if input.Tags != nil {
		params.Tags = input.Tags
	}

	if input.Metadata != nil {
		metadataJSON, marshalErr := json.Marshal(input.Metadata)
		if marshalErr != nil {
			tracing.RecordError(span, marshalErr)
			return nil, fmt.Errorf("failed to marshal metadata: %w", marshalErr)
		}
		params.Metadata = metadataJSON
	}

	if input.DueDate != nil {
		params.DueDate = pgtype.Timestamptz{
			Time:  *input.DueDate,
			Valid: true,
		}
	}

	// Database span for update
	ctx, updateSpan := tracing.StartDBSpan(ctx, "update", "todos")
	start := time.Now()
	todo, err := s.querier.UpdateTodo(ctx, params)
	duration := time.Since(start).Seconds()
	metrics.RecordDatabaseQuery("update_todo", "todos", duration, err)
	if err != nil {
		tracing.RecordError(updateSpan, err)
		updateSpan.End()
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("todo not found")
		}
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}
	updateSpan.End()

	// Record business metric
	metrics.RecordTodoOperation("update")
	return s.toTodoModel(todo), nil
}

// CompleteTodo marks a todo as completed
// Uses transaction to ensure atomicity
func (s *TodoService) CompleteTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) (*model.Todo, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "complete")
	defer span.End()
	tracing.SetAttributes(span,
		tracing.String("todo.id", todoID.String()),
		tracing.String("user.id", userID.String()),
	)

	var todo db.Todo
	var err error

	// Transaction span
	ctx, txSpan := tracing.StartSpan(ctx, tracing.TracerDB, "db.transaction")

	// Use transaction with row locking
	txErr := database.WithTransaction(ctx, s.pool.Pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
		txQuerier := db.New(tx)

		// Select for update span
		_, selectSpan := tracing.StartDBSpan(ctx, "select_for_update", "todos")
		start := time.Now()
		todo, err = txQuerier.GetTodoByIDForUpdate(ctx, todoID)
		metrics.RecordDatabaseQuery("select_for_update", "todos", time.Since(start).Seconds(), err)
		if err != nil {
			tracing.RecordError(selectSpan, err)
			selectSpan.End()
			if err == pgx.ErrNoRows {
				return fmt.Errorf("todo not found")
			}
			return fmt.Errorf("failed to get todo: %w", err)
		}
		selectSpan.End()

		// Verify ownership
		if todo.UserID != userID {
			return fmt.Errorf("unauthorized: todo does not belong to user")
		}

		// Check if already completed
		if todo.Status == db.TodoStatusCompleted {
			return fmt.Errorf("todo is already completed")
		}

		// Complete span
		_, completeSpan := tracing.StartDBSpan(ctx, "update", "todos")
		start = time.Now()
		todo, err = txQuerier.CompleteTodo(ctx, todoID)
		metrics.RecordDatabaseQuery("complete_todo", "todos", time.Since(start).Seconds(), err)
		if err != nil {
			tracing.RecordError(completeSpan, err)
			completeSpan.End()
			return fmt.Errorf("failed to complete todo: %w", err)
		}
		completeSpan.End()

		return nil
	})

	if txErr != nil {
		tracing.RecordError(txSpan, txErr)
		txSpan.End()
		return nil, txErr
	}
	txSpan.End()

	// Record business metric
	metrics.RecordTodoOperation("complete")
	return s.toTodoModel(todo), nil
}

// BulkCompleteTodos completes multiple todos
func (s *TodoService) BulkCompleteTodos(ctx context.Context, todoIDs []uuid.UUID) error {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "bulk_complete")
	defer span.End()
	tracing.SetAttributes(span, tracing.Int("todo.count", len(todoIDs)))

	// Transaction span
	ctx, txSpan := tracing.StartSpan(ctx, tracing.TracerDB, "db.transaction")

	err := database.WithTransaction(ctx, s.pool.Pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
		txQuerier := db.New(tx)

		_, dbSpan := tracing.StartDBSpan(ctx, "bulk_update", "todos")
		err := txQuerier.BulkUpdateTodoStatus(ctx, db.BulkUpdateTodoStatusParams{
			Column1: todoIDs,
			Status:  db.TodoStatusCompleted,
		})
		if err != nil {
			tracing.RecordError(dbSpan, err)
			dbSpan.End()
			return fmt.Errorf("failed to bulk complete todos: %w", err)
		}
		tracing.SetAttributes(dbSpan, tracing.Int("db.rows_affected", len(todoIDs)))
		dbSpan.End()

		return nil
	})

	if err != nil {
		tracing.RecordError(txSpan, err)
	}
	txSpan.End()

	return err
}

// BulkDeleteTodos deletes multiple todos
func (s *TodoService) BulkDeleteTodos(ctx context.Context, todoIDs []uuid.UUID, userID uuid.UUID) error {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "bulk_delete")
	defer span.End()
	tracing.SetAttributes(span,
		tracing.String("user.id", userID.String()),
		tracing.Int("todo.count", len(todoIDs)),
	)

	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "delete", "todos")
	err := s.querier.HardDeleteTodos(ctx, db.HardDeleteTodosParams{
		Column1: todoIDs,
		UserID:  userID,
	})
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		return fmt.Errorf("failed to hard delete todos: %w", err)
	}
	tracing.SetAttributes(dbSpan, tracing.Int("db.rows_affected", len(todoIDs)))
	dbSpan.End()

	return nil
}

func (s *TodoService) DeleteTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) error {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "delete")
	defer span.End()
	tracing.SetAttributes(span,
		tracing.String("todo.id", todoID.String()),
		tracing.String("user.id", userID.String()),
	)

	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "delete", "todos")
	start := time.Now()
	err := s.querier.HardDeleteTodo(ctx, db.HardDeleteTodoParams{
		ID:     todoID,
		UserID: userID,
	})
	metrics.RecordDatabaseQuery("delete", "todos", time.Since(start).Seconds(), err)
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		return fmt.Errorf("failed to hard delete todo: %w", err)
	}
	dbSpan.End()

	return nil
}

// GetTodoStats retrieves aggregated statistics
func (s *TodoService) GetTodoStats(ctx context.Context, userID uuid.UUID) (*model.TodoStats, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "get_stats")
	defer span.End()
	tracing.SetAttributes(span, tracing.String("user.id", userID.String()))

	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "select", "todos")
	stats, err := s.querier.GetTodoStats(ctx, userID)
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		return nil, fmt.Errorf("failed to get todo stats: %w", err)
	}
	dbSpan.End()

	// Add stats to span
	tracing.SetAttributes(span,
		tracing.Int64("stats.total", stats.Total),
		tracing.Int64("stats.pending", stats.Pending),
		tracing.Int64("stats.completed", stats.Completed),
	)

	return &model.TodoStats{
		Total:      stats.Total,
		Pending:    stats.Pending,
		InProgress: stats.InProgress,
		Completed:  stats.Completed,
		Overdue:    stats.Overdue,
	}, nil
}

// CountUserTodos counts total todos for a user
func (s *TodoService) CountUserTodos(ctx context.Context, userID uuid.UUID) (int64, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "count")
	defer span.End()
	tracing.SetAttributes(span, tracing.String("user.id", userID.String()))

	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "count", "todos")
	count, err := s.querier.CountUserTodos(ctx, userID)
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		return 0, fmt.Errorf("failed to count todos: %w", err)
	}
	tracing.SetAttributes(dbSpan, tracing.Int64("db.count", count))
	dbSpan.End()

	tracing.SetAttributes(span, tracing.Int64("result.count", count))
	return count, nil
}

// SearchTodosByTitle searches todos by title
func (s *TodoService) SearchTodosByTitle(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]*model.Todo, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "search")
	defer span.End()
	tracing.SetAttributes(span,
		tracing.String("user.id", userID.String()),
		tracing.String("search.query", query),
		tracing.Int("pagination.limit", int(limit)),
	)

	// validate search query
	if err := validator.ValidateSearchQuery(query); err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("invalid search query: %w", err)
	}

	// sanitize search query
	query = validator.SanitizeString(query, 100)

	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "search", "todos")
	todos, err := s.querier.SearchTodosByTitle(ctx, db.SearchTodosByTitleParams{
		UserID:  userID,
		Column2: &query,
		Limit:   limit,
	})
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		return nil, fmt.Errorf("failed to search todos: %w", err)
	}
	tracing.SetAttributes(dbSpan, tracing.Int("db.rows_affected", len(todos)))
	dbSpan.End()

	tracing.SetAttributes(span, tracing.Int("result.count", len(todos)))
	return s.toTodoModels(todos), nil
}

// GetTodosByTags retrieves todos with specific tags
func (s *TodoService) GetTodosByTags(ctx context.Context, userID uuid.UUID, tags []string) ([]*model.Todo, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "get_by_tags")
	defer span.End()
	tracing.SetAttributes(span,
		tracing.String("user.id", userID.String()),
		tracing.Int("filter.tags_count", len(tags)),
	)

	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "select", "todos")
	todos, err := s.querier.GetTodosByTags(ctx, db.GetTodosByTagsParams{
		UserID:  userID,
		Column2: tags,
	})
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		return nil, fmt.Errorf("failed to get todos by tags: %w", err)
	}
	tracing.SetAttributes(dbSpan, tracing.Int("db.rows_affected", len(todos)))
	dbSpan.End()

	tracing.SetAttributes(span, tracing.Int("result.count", len(todos)))
	return s.toTodoModels(todos), nil
}

// GetOverdueTodos retrieves overdue todos across all users (admin function)
func (s *TodoService) GetOverdueTodos(ctx context.Context, limit int32) ([]*model.Todo, error) {
	// Start service span
	ctx, span := tracing.StartServiceSpan(ctx, "todo", "get_overdue")
	defer span.End()
	tracing.SetAttributes(span, tracing.Int("pagination.limit", int(limit)))

	// Database span
	ctx, dbSpan := tracing.StartDBSpan(ctx, "select", "todos")
	results, err := s.querier.GetOverdueTodos(ctx, limit)
	if err != nil {
		tracing.RecordError(dbSpan, err)
		dbSpan.End()
		return nil, fmt.Errorf("failed to get overdue todos: %w", err)
	}
	tracing.SetAttributes(dbSpan, tracing.Int("db.rows_affected", len(results)))
	dbSpan.End()

	todos := make([]*model.Todo, len(results))
	for i, result := range results {
		todo := db.Todo{
			ID:          result.ID,
			UserID:      result.UserID,
			Title:       result.Title,
			Description: result.Description,
			Status:      result.Status,
			Priority:    result.Priority,
			Tags:        result.Tags,
			Metadata:    result.Metadata,
			DueDate:     result.DueDate,
			CompletedAt: result.CompletedAt,
			CreatedAt:   result.CreatedAt,
			UpdatedAt:   result.UpdatedAt,
		}
		todos[i] = s.toTodoModel(todo)
	}

	tracing.SetAttributes(span, tracing.Int("result.count", len(todos)))
	return todos, nil
}

// Helper methods

func (s *TodoService) validateCreateTodoInput(input model.CreateTodoInput) error {
	if input.Title == "" {
		return fmt.Errorf("title is required")
	}
	if input.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

func (s *TodoService) toTodoModel(todo db.Todo) *model.Todo {
	todoModel := &model.Todo{
		ID:        todo.ID,
		UserID:    todo.UserID,
		Title:     todo.Title,
		Status:    model.TodoStatus(todo.Status),
		Priority:  model.TodoPriority(todo.Priority),
		Tags:      todo.Tags,
		CreatedAt: todo.CreatedAt,
		UpdatedAt: todo.UpdatedAt,
	}

	// Handle nullable description
	if todo.Description != nil {
		todoModel.Description = *todo.Description
	}

	// Handle nullable due date
	if todo.DueDate.Valid {
		todoModel.DueDate = &todo.DueDate.Time
	}

	// Handle nullable completed at
	if todo.CompletedAt.Valid {
		todoModel.CompletedAt = &todo.CompletedAt.Time
	}

	// Parse metadata if present
	if len(todo.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(todo.Metadata, &metadata); err == nil {
			todoModel.Metadata = metadata
		}
	}

	return todoModel
}

func (s *TodoService) toTodoModels(todos []db.Todo) []*model.Todo {
	models := make([]*model.Todo, len(todos))
	for i, todo := range todos {
		models[i] = s.toTodoModel(todo)
	}
	return models
}
