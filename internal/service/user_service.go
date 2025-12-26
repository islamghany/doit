package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"doit/internal/data/db"
	"doit/internal/model"
	"doit/internal/tracing"
	"doit/pkg/database"
	passwordHash "doit/pkg/password_hash"
	"doit/pkg/validator"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Sentinel errors for user service
var (
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDuplicateUsername  = errors.New("username already exists")
	ErrFailedToCreateUser = errors.New("failed to create user")
)

// UserService handles all user-related business logic
// This is a FAT SERVICE - all business logic lives here
type UserService struct {
	pool    *database.Pool
	querier db.Querier
}

func NewUserService(pool *database.Pool) *UserService {
	return &UserService{
		pool:    pool,
		querier: db.New(pool),
	}
}

func NewUserServiceWithQuerier(querier db.Querier) *UserService {
	return &UserService{
		querier: querier,
	}
}

// CreateUser creates a new user with password hashing
func (s *UserService) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "CreateUser")
	defer span.End()

	tracing.SetAttributes(span,
		tracing.String("user.email", input.Email),
		tracing.String("user.username", input.Username),
		tracing.String("user.role", string(input.Role)),
	)

	// Validate input
	if err := s.validateCreateUserInput(input); err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// validate password strength
	if err := validator.ValidatePasswordStrength(input.Password, validator.DefaultPasswordStrength); err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash password
	tracing.AddEvent(span, "hashing_password")
	hashedPassword, err := passwordHash.HashPassword([]byte(input.Password))
	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Prepare metadata
	metadata := input.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create user in database
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "INSERT", "users")
	user, err := s.querier.CreateUser(dbCtx, db.CreateUserParams{
		ID:           uuid.New(),
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(hashedPassword),
		Metadata:     metadataJSON,
		Role:         string(input.Role),
	})
	dbSpan.End()

	if err != nil {
		// check if the error is a duplicate email or username
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"users_email_key\"") {
			tracing.RecordError(span, ErrDuplicateEmail)
			return nil, ErrDuplicateEmail
		}
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"users_username_key\"") {
			tracing.RecordError(span, ErrDuplicateUsername)
			return nil, ErrDuplicateUsername
		}
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	tracing.SetAttributes(span, tracing.String("user.id", user.ID.String()))
	tracing.SetOK(span)
	return toUserModel(user), nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "GetUserByID")
	defer span.End()

	tracing.SetAttributes(span, tracing.String("user.id", userID.String()))

	// Query database
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "SELECT", "users")
	user, err := s.querier.GetUserByID(dbCtx, userID)
	dbSpan.End()

	if err != nil {
		if err == pgx.ErrNoRows {
			tracing.AddEvent(span, "user_not_found")
			return nil, fmt.Errorf("user not found")
		}
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	tracing.SetOK(span)
	return toUserModel(user), nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "GetUserByEmail")
	defer span.End()

	tracing.SetAttributes(span, tracing.String("user.email", email))

	// Query database
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "SELECT", "users")
	user, err := s.querier.GetUserByEmail(dbCtx, email)
	dbSpan.End()

	if err != nil {
		if err == pgx.ErrNoRows {
			tracing.AddEvent(span, "user_not_found")
			return nil, fmt.Errorf("user not found")
		}
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	tracing.SetOK(span)
	return toUserModel(user), nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "GetUserByUsername")
	defer span.End()

	tracing.SetAttributes(span, tracing.String("user.username", username))

	// Query database
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "SELECT", "users")
	user, err := s.querier.GetUserByUsername(dbCtx, username)
	dbSpan.End()

	if err != nil {
		if err == pgx.ErrNoRows {
			tracing.AddEvent(span, "user_not_found")
			return nil, fmt.Errorf("user not found")
		}
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	tracing.SetOK(span)
	return toUserModel(user), nil
}

// AuthenticateUser validates credentials and updates last login
func (s *UserService) AuthenticateUser(ctx context.Context, input model.LoginInput) (*model.User, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "AuthenticateUser")
	defer span.End()

	tracing.SetAttributes(span, tracing.String("user.email", input.Email))

	// Get user by email
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "SELECT", "users")
	user, err := s.querier.GetUserByEmail(dbCtx, input.Email)
	dbSpan.End()

	if err != nil {
		if err == pgx.ErrNoRows {
			tracing.AddEvent(span, "authentication_failed", tracing.String("reason", "user_not_found"))
			return nil, ErrInvalidCredentials
		}
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// // Check if user is active
	// if !user.IsActive {
	// 	return nil, fmt.Errorf("user account is inactive")
	// }

	// Verify password
	tracing.AddEvent(span, "verifying_password")
	if err := passwordHash.ComparePassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		tracing.AddEvent(span, "authentication_failed", tracing.String("reason", "invalid_password"))
		return nil, ErrInvalidCredentials
	}

	// Update last login time (non-critical, don't fail auth if this fails)
	updateCtx, updateSpan := tracing.StartDBSpan(ctx, "UPDATE", "users")
	if err := s.querier.UpdateUserLastLogin(updateCtx, user.ID); err != nil {
		// Log error but don't fail authentication
		tracing.AddEvent(span, "last_login_update_failed", tracing.String("error", err.Error()))
		fmt.Printf("failed to update last login: %v\n", err)
	}
	updateSpan.End()

	tracing.SetAttributes(span, tracing.String("user.id", user.ID.String()))
	tracing.AddEvent(span, "authentication_successful")
	tracing.SetOK(span)
	return toUserModel(user), nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, input model.UpdateUserInput) (*model.User, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "UpdateUser")
	defer span.End()

	tracing.SetAttributes(span, tracing.String("user.id", userID.String()))

	// Prepare metadata if provided
	var metadataJSON []byte
	if input.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(input.Metadata)
		if err != nil {
			tracing.RecordError(span, err)
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Update user in database
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "UPDATE", "users")
	user, err := s.querier.UpdateUser(dbCtx, db.UpdateUserParams{
		ID:       userID,
		Email:    input.Email,
		Username: input.Username,
		IsActive: input.IsActive,
		Metadata: metadataJSON,
	})
	dbSpan.End()

	if err != nil {
		if err == pgx.ErrNoRows {
			tracing.AddEvent(span, "user_not_found")
			return nil, fmt.Errorf("user not found")
		}
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	tracing.SetOK(span)
	return toUserModel(user), nil
}

// UpdateUserPassword updates a user's password
func (s *UserService) UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "UpdateUserPassword")
	defer span.End()

	tracing.SetAttributes(span, tracing.String("user.id", userID.String()))

	// Hash new password
	tracing.AddEvent(span, "hashing_password")
	hashedPassword, err := passwordHash.HashPassword([]byte(newPassword))
	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password in database
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "UPDATE", "users")
	err = s.querier.UpdateUserPassword(dbCtx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: string(hashedPassword),
	})
	dbSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	tracing.SetOK(span)
	return nil
}

// VerifyUserEmail marks a user's email as verified
func (s *UserService) VerifyUserEmail(ctx context.Context, userID uuid.UUID) error {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "VerifyUserEmail")
	defer span.End()

	tracing.SetAttributes(span, tracing.String("user.id", userID.String()))

	// Update email verification status
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "UPDATE", "users")
	err := s.querier.VerifyUserEmail(dbCtx, userID)
	dbSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to verify email: %w", err)
	}

	tracing.AddEvent(span, "email_verified")
	tracing.SetOK(span)
	return nil
}

// ListUsers retrieves paginated list of users
func (s *UserService) ListUsers(ctx context.Context, filter model.UserFilter) ([]*model.User, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "ListUsers")
	defer span.End()

	tracing.SetAttributes(span,
		tracing.Int("pagination.limit", int(filter.Limit)),
		tracing.Int("pagination.offset", int(filter.Offset)),
	)

	// If email filter is provided, use search
	if filter.Email != nil && *filter.Email != "" {
		tracing.SetAttributes(span, tracing.String("filter.email", *filter.Email))

		dbCtx, dbSpan := tracing.StartDBSpan(ctx, "SELECT", "users")
		tracing.SetAttributes(dbSpan, tracing.String("db.query_type", "search_by_email"))
		users, err := s.querier.SearchUsersByEmail(dbCtx, db.SearchUsersByEmailParams{
			Column1: filter.Email,
			Limit:   filter.Limit,
		})
		dbSpan.End()

		if err != nil {
			tracing.RecordError(span, err)
			return nil, fmt.Errorf("failed to search users: %w", err)
		}

		tracing.SetAttributes(span, tracing.Int("result.count", len(users)))
		tracing.SetOK(span)
		return s.toUserModels(users), nil
	}

	// Otherwise use regular list
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "SELECT", "users")
	tracing.SetAttributes(dbSpan, tracing.String("db.query_type", "list"))
	users, err := s.querier.ListUsers(dbCtx, db.ListUsersParams{
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
	dbSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	tracing.SetAttributes(span, tracing.Int("result.count", len(users)))
	tracing.SetOK(span)
	return s.toUserModels(users), nil
}

// CountUsers returns the total count of users
func (s *UserService) CountUsers(ctx context.Context) (int64, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "CountUsers")
	defer span.End()

	// Query database
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "SELECT", "users")
	tracing.SetAttributes(dbSpan, tracing.String("db.query_type", "count"))
	count, err := s.querier.CountUsers(dbCtx)
	dbSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	tracing.SetAttributes(span, tracing.Int64("result.count", count))
	tracing.SetOK(span)
	return count, nil
}

// BatchUpdateUsersMetadata updates metadata for multiple users
func (s *UserService) BatchUpdateUsersMetadata(ctx context.Context, userIDs []uuid.UUID, metadata map[string]interface{}) error {
	ctx, span := tracing.StartServiceSpan(ctx, "UserService", "BatchUpdateUsersMetadata")
	defer span.End()

	tracing.SetAttributes(span, tracing.Int("user.count", len(userIDs)))

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Batch update in database
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "UPDATE", "users")
	tracing.SetAttributes(dbSpan,
		tracing.String("db.query_type", "bulk_update"),
		tracing.Int("db.batch_size", len(userIDs)),
	)
	err = s.querier.BulkUpdateUsersMetadata(dbCtx, db.BulkUpdateUsersMetadataParams{
		Column1: userIDs,
		Column2: metadataJSON,
	})
	dbSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("failed to batch update users: %w", err)
	}

	tracing.SetOK(span)
	return nil
}

// Helper methods

func (s *UserService) validateCreateUserInput(input model.CreateUserInput) error {
	if input.Email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	if input.Username == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidInput)
	}
	if len(input.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidInput)
	}
	return nil
}

func toUserModel(user db.User) *model.User {
	userModel := &model.User{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		EmailVerified: user.EmailVerified,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		TokenVersion:  *user.TokenVersion,
		Role:          model.UserRole(user.Role),
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
		models[i] = toUserModel(user)
	}
	return models
}
