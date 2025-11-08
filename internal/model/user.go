// Package model contains domain models and business entities.
package model

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user (RBAC)
type UserRole string

const (
	UserRoleUser      UserRole = "user"      // Regular user
	UserRoleAdmin     UserRole = "admin"     // Full system access
	UserRoleModerator UserRole = "moderator" // Limited admin access
)

// User represents the domain model for a user
// @Description User account information (password field is never included in responses)
type User struct {
	// Unique user identifier (UUID v4)
	ID uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// User's email address
	Email string `json:"email" example:"john.doe@example.com"`

	// User's unique username
	Username string `json:"username" example:"johndoe"`

	// User's role (user, admin, moderator)
	Role UserRole `json:"role" example:"user" enums:"user,admin,moderator"`

	// Whether the user's email has been verified
	EmailVerified bool `json:"email_verified" example:"false"`

	// Whether the user account is active
	IsActive bool `json:"is_active" example:"true"`

	// Additional metadata (flexible key-value storage)
	Metadata map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`

	// Last successful login timestamp
	LastLoginAt *time.Time `json:"last_login_at,omitempty" example:"2025-11-08T10:30:00Z"`

	// Account creation timestamp
	CreatedAt time.Time `json:"created_at" example:"2025-11-08T10:00:00Z"`

	// Last update timestamp
	UpdatedAt time.Time `json:"updated_at" example:"2025-11-08T10:00:00Z"`

	// Token version for refresh token rotation (used internally for security)
	TokenVersion int32 `json:"token_version,omitempty" example:"1"`
}

func (u *User) IsUserActive() bool {
	return u.IsActive
}

func (u *User) IsUserEmailVerified() bool {
	return u.EmailVerified
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Email    string                 `json:"email"`
	Username string                 `json:"username"`
	Password string                 `json:"password"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Role     UserRole               `json:"role"`
}

// UpdateUserInput represents input for updating a user
type UpdateUserInput struct {
	Email    *string                `json:"email,omitempty"`
	Username *string                `json:"username,omitempty"`
	IsActive *bool                  `json:"is_active,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// LoginInput represents credentials for authentication
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserFilter represents filtering options for listing users
type UserFilter struct {
	Email  *string `json:"email,omitempty"`
	Limit  int32   `json:"limit"`
	Offset int32   `json:"offset"`
}

// ================================
// User Context

type userContextKey string

const (
	UserContextKey userContextKey = "user"
)

func SetUserContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

func GetUserContext(ctx context.Context) *User {
	user, ok := ctx.Value(UserContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// GetUserFromContext retrieves user from context and returns error if not found
func GetUserFromContext(ctx context.Context) (*User, error) {
	user := GetUserContext(ctx)
	if user == nil {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}
