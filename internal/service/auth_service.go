package service

import (
	"context"
	"doit/internal/data/db"
	"doit/internal/model"
	"doit/pkg/database"
	"doit/pkg/logger"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	pool    *database.Pool
	querier db.Querier
	log *logger.Logger
}

// Sentinel errors for auth service
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
)

func NewAuthService(pool *database.Pool, log *logger.Logger) *AuthService {
	return &AuthService{pool: pool, querier: db.New(pool), log: log}
}

func (s *AuthService) AuthenticateUser(ctx context.Context, input model.LoginInput) (*model.User, error) {
	user, err := s.querier.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrInvalidCredentials // Don't reveal if user exists for security
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}
	
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	
	// Update last login (non-critical, don't fail auth if this fails)
	if err := s.querier.UpdateUserLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail authentication
		s.log.Error(ctx, "warning: failed to update last login", "error", err)
	}

	return toUserModel(user), nil
}
