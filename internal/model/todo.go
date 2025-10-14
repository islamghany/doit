package model

import (
	"time"

	"github.com/google/uuid"
)

// TodoStatus represents the status of a todo
type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"
	TodoStatusArchived   TodoStatus = "archived"
)

// TodoPriority represents the priority of a todo
type TodoPriority string

const (
	TodoPriorityLow    TodoPriority = "low"
	TodoPriorityMedium TodoPriority = "medium"
	TodoPriorityHigh   TodoPriority = "high"
	TodoPriorityUrgent TodoPriority = "urgent"
)

// Todo represents the domain model for a todo
type Todo struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Status      TodoStatus             `json:"status"`
	Priority    TodoPriority           `json:"priority"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CreateTodoInput represents input for creating a todo
type CreateTodoInput struct {
	UserID      uuid.UUID              `json:"user_id" validate:"required"`
	Title       string                 `json:"title" validate:"required,min=1,max=255"`
	Description string                 `json:"description,omitempty"`
	Priority    TodoPriority           `json:"priority" validate:"omitempty,oneof=low medium high urgent"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
}

// UpdateTodoInput represents input for updating a todo
type UpdateTodoInput struct {
	Title       *string                `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string                `json:"description,omitempty"`
	Status      *TodoStatus            `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress completed archived"`
	Priority    *TodoPriority          `json:"priority,omitempty" validate:"omitempty,oneof=low medium high urgent"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
}

// TodoFilter represents filtering options for listing todos
type TodoFilter struct {
	UserID    uuid.UUID     `json:"user_id" validate:"required"`
	Status    *TodoStatus   `json:"status,omitempty"`
	Priority  *TodoPriority `json:"priority,omitempty"`
	Tags      []string      `json:"tags,omitempty"`
	DueBefore *time.Time    `json:"due_before,omitempty"`
	Limit     int32         `json:"limit" validate:"min=1,max=100"`
	Offset    int32         `json:"offset" validate:"min=0"`
}

// TodoStats represents aggregated statistics for todos
type TodoStats struct {
	Total      int64 `json:"total"`
	Pending    int64 `json:"pending"`
	InProgress int64 `json:"in_progress"`
	Completed  int64 `json:"completed"`
	Overdue    int64 `json:"overdue"`
}
