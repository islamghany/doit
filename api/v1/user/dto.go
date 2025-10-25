package user

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email    string                 `json:"email" validate:"required,email"`
	Username string                 `json:"username" validate:"required,min=3,max=50"`
	Password string                 `json:"password" validate:"required,min=8"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Email    *string                `json:"email,omitempty" validate:"omitempty,email"`
	Username *string                `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	IsActive *bool                  `json:"is_active,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UserResponse represents the response for user operations
type UserResponse struct {
	ID            string                 `json:"id"`
	Email         string                 `json:"email"`
	Username      string                 `json:"username"`
	EmailVerified bool                   `json:"email_verified"`
	IsActive      bool                   `json:"is_active"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}
