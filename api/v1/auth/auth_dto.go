package auth

import "doit/internal/model"

// LoginInput represents user login credentials
// @Description User login request payload
type LoginInput struct {
	// User's email address
	Email string `json:"email" validate:"required,email" example:"john.doe@example.com"`

	// User's password
	Password string `json:"password" validate:"required" example:"SecurePass123!"`
}

// RegisterInput represents user registration data
// @Description User registration request payload
type RegisterInput struct {
	// User's email address (must be unique)
	Email string `json:"email" validate:"required,email" example:"jane.smith@example.com"`

	// Unique username (max 50 characters)
	Username string `json:"username" validate:"required,max=50" example:"janesmith"`

	// Password (minimum 8 characters, should include uppercase, lowercase, numbers, and special characters)
	Password string `json:"password" validate:"required,min=8" example:"SecurePass123!"`

	// Password confirmation (must match password field)
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password" example:"SecurePass123!"`

	// User role: user (default), admin, or moderator
	Role string `json:"role" validate:"required,oneof=user admin moderator" example:"user" enums:"user,admin,moderator"`
}

// LoginResponse represents successful login response
// @Description User login response with tokens
type LoginResponse struct {
	User   interface{}     `json:"user" swaggertype:"object"`
	Tokens model.TokenPair `json:"tokens"`
}

// RefreshResponse represents refresh token response
// @Description Response containing new access token
type RefreshResponse struct {
	// New JWT access token
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// New refresh token (rotation)
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// Token type (always "Bearer")
	TokenType string `json:"token_type" example:"Bearer"`

	// Access token expiration time in seconds
	ExpiresIn int64 `json:"expires_in" example:"900"`
}

// ErrorResponse represents error response
// @Description Standard error response format
type ErrorResponse struct {
	// Error message describing what went wrong
	Error string `json:"error" example:"validation error"`

	// HTTP status code
	Code int `json:"code,omitempty" example:"400"`

	// Field-specific validation errors (optional)
	Fields map[string][]string `json:"fields,omitempty" swaggertype:"object"`
}

// SuccessResponse represents generic success response
// @Description Generic success response
type SuccessResponse struct {
	// Success message
	Message string `json:"message" example:"operation completed successfully"`

	// Optional response data
	Data interface{} `json:"data,omitempty" swaggertype:"object"`
}

// MessageResponse represents simple message response
// @Description Simple message response (used for logout, delete, etc.)
type MessageResponse struct {
	// Response message
	Message string `json:"message" example:"logged out successfully"`
}
