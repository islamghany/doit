package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"doit/internal/data/db"
	"doit/internal/model"
	"doit/internal/token"
	"doit/pkg/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Sentinel errors for token service
var (
	ErrSecurityAlert = errors.New("security alert")
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
)

// TokenService manages token lifecycle (creation, verification, revocation)
type TokenService struct {
	tokenMaker           token.TokenMaker
	pool                 *database.Pool
	querier              db.Querier
	accessTokenDuration  int
	refreshTokenDuration int
}

func NewTokenService(
	pool *database.Pool,
	tokenMaker token.TokenMaker,
	accessTokenDuration int, refreshTokenDuration int,
) *TokenService {
	return &TokenService{
		tokenMaker:           tokenMaker,
		pool:                 pool,
		querier:              db.New(pool),
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

// CreateTokenPair creates both access and refresh tokens for a user
func (s *TokenService) CreateTokenPair(ctx context.Context, user model.User, deviceInfo model.DeviceInfo) (*model.TokenPair, error) {
	// 1. Create short-lived access token (stateless)
	accessToken, _, err := s.tokenMaker.CreateToken(token.TokenParams{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Version:  int(user.TokenVersion),
		Duration: time.Duration(s.accessTokenDuration) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// 2. Create long-lived refresh token (stateful)
	refreshTokenString, refreshPayload, err := s.tokenMaker.CreateToken(token.TokenParams{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Version:  int(user.TokenVersion),
		Duration: time.Duration(s.refreshTokenDuration) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	// 3. Hash refresh token before storing (never store plaintext!)
	tokenHash := hashToken(refreshTokenString)

	// 4. Prepare device info as JSON
	deviceJSON, err := json.Marshal(deviceInfo)
	if err != nil {
		return nil, fmt.Errorf("marshal device info: %w", err)
	}

	// 5. Store refresh token in database
	_, err = s.querier.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		ID:        refreshPayload.ID,
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamp{
			Time:  refreshPayload.ExpiredAt,
			Valid: true,
		},
		DeviceInfo: deviceJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    s.accessTokenDuration,
	}, nil
}

// VerifyAccessToken verifies an access token and returns the payload
// This is fast - only checks JWT signature + version (no DB lookup of token itself)
func (s *TokenService) VerifyAccessToken(ctx context.Context, tokenString string) (*token.Payload, error) {
	// 1. Verify JWT signature and expiration
	payload, err := s.tokenMaker.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 2. Check token version (handles password changes, security events)
	currentVersion, err := s.querier.GetUserTokenVersion(ctx, payload.UserID)
	if err != nil {
		return nil, fmt.Errorf("get token version: %w", err)
	}

	// 3. Version mismatch = token invalidated (password changed, security event)
	if payload.Version != int(*currentVersion) {
		return nil, token.ErrInvalidToken
	}

	return payload, nil
}

// RefreshAccessToken exchanges a refresh token for a new access token
func (s *TokenService) RefreshAccessToken(ctx context.Context, refreshTokenString string) (*model.TokenPair, error) {
	// 1. Verify refresh token JWT signature and expiration
	payload, err := s.tokenMaker.VerifyToken(refreshTokenString)
	if err != nil {
		return nil, token.ErrInvalidToken
	}

	// 2. Hash the token to lookup in database
	tokenHash := hashToken(refreshTokenString)

	// 3. Get refresh token from database (including revoked ones for security check)
	storedToken, err := s.querier.GetRefreshTokenIncludingRevoked(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}
	// 4. SECURITY CHECK: Detect token reuse (revoked token being used)
	if storedToken.IsRevoked {
		// 🚨 CRITICAL SECURITY EVENT!
		// Someone is trying to use a revoked token - possible theft/replay attack
		// Log the security incident
		fmt.Printf("🚨 SECURITY ALERT: Revoked token reuse detected!\n")
		fmt.Printf("   User ID: %s\n", storedToken.UserID)
		fmt.Printf("   Token ID: %s\n", storedToken.ID)
		fmt.Printf("   Revoked At: %s\n", storedToken.LastUsedAt.Time)
		fmt.Printf("   Time Since Revoke: %s\n", time.Since(storedToken.LastUsedAt.Time))

		// Security Response: Revoke ALL user tokens immediately
		err = s.RevokeAllUserTokens(ctx, storedToken.UserID)
		if err != nil {
			fmt.Printf("Failed to revoke all tokens: %v\n", err)
		}

		// TODO: Send security alert email/notification to user
		// s.notificationService.SendSecurityAlert(storedToken.UserID, ...)

		return nil, ErrSecurityAlert
	}
	// 5. Check if token is expired
	if time.Now().After(storedToken.ExpiresAt.Time) {
		return nil, token.ErrExpiredToken
	}
	// 6. Update last used timestamp (for session tracking)
	_ = s.querier.UpdateRefreshTokenUsage(ctx, storedToken.ID)

	// 7. Get current user data (fresh from DB)
	user, err := s.querier.GetUserByID(ctx, payload.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 8. Check if user is still active (not banned/deleted)
	// TODO: Add user status check if you have user.IsActive field
	// if !user.IsActive {
	//     return nil, errors.New("user account disabled")
	// }

	// 9. Create new access token with current user data and version
	accessToken, _, err := s.tokenMaker.CreateToken(token.TokenParams{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Version:  int(*user.TokenVersion),
		Duration: 15 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("create access token: %w", err)
	}
	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    s.accessTokenDuration,
	}, nil
}

// Helper: Hash token using SHA256 before storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Logout revokes a specific refresh token (single device logout)
func (s *TokenService) Logout(ctx context.Context, refreshTokenString string) error {
	tokenHash := hashToken(refreshTokenString)
	return s.querier.RevokeRefreshToken(ctx, tokenHash)
}

// RevokeAllUserTokens revokes all tokens for a user (logout all devices)
// Use when: password change, security breach, admin action
func (s *TokenService) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	// 1. Increment token version (invalidates ALL access tokens immediately)
	_, err := s.querier.IncrementUserTokenVersion(ctx, userID)
	if err != nil {
		return fmt.Errorf("increment token version: %w", err)
	}

	// 2. Revoke all refresh tokens (logout all devices)
	err = s.querier.RevokeAllUserRefreshTokens(ctx, userID)
	if err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}

	return nil
}

// GetUserSessions returns all active sessions for a user
func (s *TokenService) GetUserSessions(ctx context.Context, userID uuid.UUID, currentTokenID uuid.UUID) ([]model.Session, error) {
	tokens, err := s.querier.GetUserActiveRefreshTokens(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get active tokens: %w", err)
	}

	sessions := make([]model.Session, len(tokens))
	for i, t := range tokens {
		var deviceInfo map[string]interface{}
		if len(t.DeviceInfo) > 0 {
			_ = json.Unmarshal(t.DeviceInfo, &deviceInfo)
		}

		sessions[i] = model.Session{
			ID:         t.ID,
			CreatedAt:  t.CreatedAt.Time,
			LastUsedAt: t.LastUsedAt.Time,
			DeviceInfo: deviceInfo,
			IsCurrent:  t.ID == currentTokenID,
		}
	}

	return sessions, nil
}
