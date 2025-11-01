package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"doit/internal/data/db"
	"doit/internal/model"
	"doit/pkg/database"

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
	if err := s.validateCreateTodoInput(input); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

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
	if err != nil {
		return nil, ErrTodoNotFound
	}

	return s.toTodoModel(todo), nil
}

// GetTodoByID retrieves a todo by ID
func (s *TodoService) GetTodoByID(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) (*model.Todo, error) {
	todo, err := s.querier.GetTodoByID(ctx, todoID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	// Verify ownership
	if err = VerifyOwnership(todo.UserID, userID, "todo"); err != nil {
		return nil, err
	}

	return s.toTodoModel(todo), nil
}

// ListUserTodos retrieves paginated todos for a user
func (s *TodoService) ListUserTodos(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*model.Todo, error) {
	todos, err := s.querier.ListTodosByUser(ctx, db.ListTodosByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list todos: %w", err)
	}

	return s.toTodoModels(todos), nil
}

// ListTodosByStatus retrieves todos filtered by status
func (s *TodoService) ListTodosByStatus(ctx context.Context, userID uuid.UUID, status model.TodoStatus, limit, offset int32) ([]*model.Todo, error) {
	todos, err := s.querier.ListTodosByUserAndStatus(ctx, db.ListTodosByUserAndStatusParams{
		UserID: userID,
		Status: db.TodoStatus(status),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list todos by status: %w", err)
	}

	return s.toTodoModels(todos), nil
}

// UpdateTodo updates a todo
func (s *TodoService) UpdateTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID, input model.UpdateTodoInput) (*model.Todo, error) {
	// First, verify ownership
	existingTodo, err := s.querier.GetTodoByID(ctx, todoID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	if err = VerifyOwnership(existingTodo.UserID, userID, "todo"); err != nil {
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
		metadataJSON, err := json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		params.Metadata = metadataJSON
	}

	if input.DueDate != nil {
		params.DueDate = pgtype.Timestamptz{
			Time:  *input.DueDate,
			Valid: true,
		}
	}

	todo, err := s.querier.UpdateTodo(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("todo not found")
		}
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return s.toTodoModel(todo), nil
}

// CompleteTodo marks a todo as completed
// Uses transaction to ensure atomicity
func (s *TodoService) CompleteTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) (*model.Todo, error) {
	var todo db.Todo
	var err error

	// Use transaction with row locking
	txErr := database.WithTransaction(ctx, s.pool.Pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
		txQuerier := db.New(tx)

		// Lock the row for update
		todo, err = txQuerier.GetTodoByIDForUpdate(ctx, todoID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return fmt.Errorf("todo not found")
			}
			return fmt.Errorf("failed to get todo: %w", err)
		}

		// Verify ownership
		if todo.UserID != userID {
			return fmt.Errorf("unauthorized: todo does not belong to user")
		}

		// Check if already completed
		if todo.Status == db.TodoStatusCompleted {
			return fmt.Errorf("todo is already completed")
		}

		// Complete the todo
		todo, err = txQuerier.CompleteTodo(ctx, todoID)
		if err != nil {
			return fmt.Errorf("failed to complete todo: %w", err)
		}

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return s.toTodoModel(todo), nil
}

// BulkCompleteTodos completes multiple todos
func (s *TodoService) BulkCompleteTodos(ctx context.Context, todoIDs []uuid.UUID) error {
	return database.WithTransaction(ctx, s.pool.Pool, database.DefaultTxOptions(), func(tx pgx.Tx) error {
		txQuerier := db.New(tx)

		err := txQuerier.BulkUpdateTodoStatus(ctx, db.BulkUpdateTodoStatusParams{
			Column1: todoIDs,
			Status:  db.TodoStatusCompleted,
		})
		if err != nil {
			return fmt.Errorf("failed to bulk complete todos: %w", err)
		}

		return nil
	})
}

// BulkDeleteTodos deletes multiple todos
func (s *TodoService) BulkDeleteTodos(ctx context.Context, todoIDs []uuid.UUID, userID uuid.UUID) error {
	err := s.querier.HardDeleteTodos(ctx, db.HardDeleteTodosParams{
		Column1: todoIDs,
		UserID:  userID,
	})
	if err != nil {
		return fmt.Errorf("failed to hard delete todos: %w", err)
	}
	return nil
}

func (s *TodoService) DeleteTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) error {
	err := s.querier.HardDeleteTodo(ctx, db.HardDeleteTodoParams{
		ID:     todoID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to hard delete todo: %w", err)
	}
	return nil
}

// GetTodoStats retrieves aggregated statistics
func (s *TodoService) GetTodoStats(ctx context.Context, userID uuid.UUID) (*model.TodoStats, error) {
	stats, err := s.querier.GetTodoStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo stats: %w", err)
	}

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
	count, err := s.querier.CountUserTodos(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count todos: %w", err)
	}
	return count, nil
}

// SearchTodosByTitle searches todos by title
func (s *TodoService) SearchTodosByTitle(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]*model.Todo, error) {
	todos, err := s.querier.SearchTodosByTitle(ctx, db.SearchTodosByTitleParams{
		UserID:  userID,
		Column2: &query,
		Limit:   limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search todos: %w", err)
	}

	return s.toTodoModels(todos), nil
}

// GetTodosByTags retrieves todos with specific tags
func (s *TodoService) GetTodosByTags(ctx context.Context, userID uuid.UUID, tags []string) ([]*model.Todo, error) {
	todos, err := s.querier.GetTodosByTags(ctx, db.GetTodosByTagsParams{
		UserID:  userID,
		Column2: tags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get todos by tags: %w", err)
	}

	return s.toTodoModels(todos), nil
}

// GetOverdueTodos retrieves overdue todos across all users (admin function)
func (s *TodoService) GetOverdueTodos(ctx context.Context, limit int32) ([]*model.Todo, error) {
	results, err := s.querier.GetOverdueTodos(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue todos: %w", err)
	}

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
