package todo

import "time"

// CreateTodoInput represents the input for creating a todo
// @Description Create todo request payload
type CreateTodoInput struct {
	// Todo title
	Title string `json:"title" validate:"required,max=255" example:"Buy groceries"`

	// Todo description
	Description string `json:"description" validate:"omitempty,max=1000" example:"Buy groceries for the week"`

	// Todo priority
	Priority string `json:"priority" validate:"omitempty,oneof=low medium high urgent" example:"medium"`

	// Todo tags
	Tags []string `json:"tags" validate:"omitempty,max=10" example:"groceries,shopping"`

	// Todo metadata
	Metadata map[string]interface{} `json:"metadata" validate:"omitempty" example:"{\"category\":\"shopping\"}"`

	// Todo due date
	DueDate *time.Time `json:"due_date" validate:"omitempty,datetime" example:"2025-01-01T00:00:00Z"`
}
