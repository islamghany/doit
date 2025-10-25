package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents the domain model for a user
type User struct {
	ID            uuid.UUID              `json:"id"`
	Email         string                 `json:"email"`
	Username      string                 `json:"username"`
	EmailVerified bool                   `json:"email_verified"`
	IsActive      bool                   `json:"is_active"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	LastLoginAt   *time.Time             `json:"last_login_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

func (u *User) IsUserActive() bool {
	return u.IsActive
}

func (u *User) IsUserEmailVerified() bool {
	return u.EmailVerified
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Email    string                 `json:"email" validate:"required,email"`
	Username string                 `json:"username" validate:"required,min=3,max=50"`
	Password string                 `json:"password" validate:"required,min=8"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserInput represents input for updating a user
type UpdateUserInput struct {
	Email    *string                `json:"email,omitempty" validate:"omitempty,email"`
	Username *string                `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	IsActive *bool                  `json:"is_active,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// LoginInput represents credentials for authentication
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserFilter represents filtering options for listing users
type UserFilter struct {
	Email  *string `json:"email,omitempty"`
	Limit  int32   `json:"limit" validate:"min=1,max=100"`
	Offset int32   `json:"offset" validate:"min=0"`
}
