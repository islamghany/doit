package service

import (
	"context"
	"doit/internal/data/db"
	"doit/internal/model"
	"doit/pkg/database"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles all user-related business logic
// This is a FAT SERVICE - all business logic lives here
type UserService struct {
	pool    *database.Pool
	queries *db.Queries
}

func NewUserService(pool *database.Pool) *UserService {
	return &UserService{
		pool:    pool,
		queries: db.New(pool),
	}
}

// CreateUser creates a new user with password hashing
func (s *UserService) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	// Validate input
	if err := s.validateCreateUserInput(input); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
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

	// Create user
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		ID:           uuid.New(),
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(hashedPassword),
		Metadata:     metadataJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.toUserModel(user), nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toUserModel(user), nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toUserModel(user), nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := s.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toUserModel(user), nil
}

// AuthenticateUser validates credentials and updates last login
func (s *UserService) AuthenticateUser(ctx context.Context, input model.LoginInput) (*model.User, error) {
	// Get user by email
	user, err := s.queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login time (non-critical, don't fail auth if this fails)
	if err := s.queries.UpdateUserLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail authentication
		fmt.Printf("failed to update last login: %v\n", err)
	}

	return s.toUserModel(user), nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, input model.UpdateUserInput) (*model.User, error) {
	// Prepare metadata if provided
	var metadataJSON []byte
	if input.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	user, err := s.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:       userID,
		Email:    input.Email,
		Username: input.Username,
		IsActive: input.IsActive,
		Metadata: metadataJSON,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.toUserModel(user), nil
}

// UpdateUserPassword updates a user's password
func (s *UserService) UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = s.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// VerifyUserEmail marks a user's email as verified
func (s *UserService) VerifyUserEmail(ctx context.Context, userID uuid.UUID) error {
	err := s.queries.VerifyUserEmail(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}
	return nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	err := s.queries.SoftDeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers retrieves paginated list of users
func (s *UserService) ListUsers(ctx context.Context, filter model.UserFilter) ([]*model.User, error) {
	// If email filter is provided, use search
	if filter.Email != nil && *filter.Email != "" {
		users, err := s.queries.SearchUsersByEmail(ctx, db.SearchUsersByEmailParams{
			Column1: filter.Email,
			Limit:   filter.Limit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to search users: %w", err)
		}
		return s.toUserModels(users), nil
	}

	// Otherwise use regular list
	users, err := s.queries.ListUsers(ctx, db.ListUsersParams{
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return s.toUserModels(users), nil
}

// CountUsers returns the total count of users
func (s *UserService) CountUsers(ctx context.Context) (int64, error) {
	count, err := s.queries.CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// BatchUpdateUsersMetadata updates metadata for multiple users
func (s *UserService) BatchUpdateUsersMetadata(ctx context.Context, userIDs []uuid.UUID, metadata map[string]interface{}) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = s.queries.BulkUpdateUsersMetadata(ctx, db.BulkUpdateUsersMetadataParams{
		Column1: userIDs,
		Column2: metadataJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to batch update users: %w", err)
	}

	return nil
}

// Helper methods

func (s *UserService) validateCreateUserInput(input model.CreateUserInput) error {
	if input.Email == "" {
		return fmt.Errorf("email is required")
	}
	if input.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(input.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

func (s *UserService) toUserModel(user db.User) *model.User {
	userModel := &model.User{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		EmailVerified: user.EmailVerified,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	// Handle nullable last login
	if user.LastLoginAt.Valid {
		userModel.LastLoginAt = &user.LastLoginAt.Time
	}

	// Parse metadata if present
	if len(user.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(user.Metadata, &metadata); err == nil {
			userModel.Metadata = metadata
		}
	}

	return userModel
}

func (s *UserService) toUserModels(users []db.User) []*model.User {
	models := make([]*model.User, len(users))
	for i, user := range users {
		models[i] = s.toUserModel(user)
	}
	return models
}
