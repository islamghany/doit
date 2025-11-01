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
type User struct {
	ID            uuid.UUID              `json:"id"`
	Email         string                 `json:"email"`
	Username      string                 `json:"username"`
	Role          UserRole               `json:"role"`
	EmailVerified bool                   `json:"email_verified"`
	IsActive      bool                   `json:"is_active"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	LastLoginAt   *time.Time             `json:"last_login_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	TokenVersion  int32                  `json:"token_version,omitempty"`
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
