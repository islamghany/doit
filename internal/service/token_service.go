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
	"doit/internal/tracing"
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
	ctx, span := tracing.StartServiceSpan(ctx, "TokenService", "CreateTokenPair")
	defer span.End()

	tracing.SetAttributes(span,
		tracing.String("user.id", user.ID.String()),
		tracing.String("user.email", user.Email),
	)

	// 1. Create short-lived access token (stateless)
	tracing.AddEvent(span, "creating_access_token")
	accessToken, _, err := s.tokenMaker.CreateToken(token.TokenParams{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     string(user.Role),
		Version:  int(user.TokenVersion),
		Duration: time.Duration(s.accessTokenDuration) * time.Second,
	})
	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// 2. Create long-lived refresh token (stateful)
	tracing.AddEvent(span, "creating_refresh_token")
	refreshTokenString, refreshPayload, err := s.tokenMaker.CreateToken(token.TokenParams{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     string(user.Role),
		Version:  int(user.TokenVersion),
		Duration: time.Duration(s.refreshTokenDuration) * time.Second,
	})
	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	// 3. Hash refresh token before storing (never store plaintext!)
	tokenHash := hashToken(refreshTokenString)

	// 4. Prepare device info as JSON
	deviceJSON, err := json.Marshal(deviceInfo)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("marshal device info: %w", err)
	}

	// 5. Store refresh token in database
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "INSERT", "refresh_tokens")
	_, err = s.querier.CreateRefreshToken(dbCtx, db.CreateRefreshTokenParams{
		ID:        refreshPayload.ID,
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamp{
			Time:  refreshPayload.ExpiredAt,
			Valid: true,
		},
		DeviceInfo: deviceJSON,
	})
	dbSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	tracing.AddEvent(span, "token_pair_created")
	tracing.SetOK(span)
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
	ctx, span := tracing.StartServiceSpan(ctx, "TokenService", "VerifyAccessToken")
	defer span.End()

	// 1. Verify JWT signature and expiration
	tracing.AddEvent(span, "verifying_jwt_signature")
	payload, err := s.tokenMaker.VerifyToken(tokenString)
	if err != nil {
		tracing.AddEvent(span, "jwt_verification_failed", tracing.String("error", err.Error()))
		tracing.RecordError(span, err)
		return nil, err
	}

	tracing.SetAttributes(span,
		tracing.String("user.id", payload.UserID.String()),
		tracing.String("token.id", payload.ID.String()),
	)

	// 2. Check token version (handles password changes, security events)
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "SELECT", "users")
	tracing.SetAttributes(dbSpan, tracing.String("db.query_type", "get_token_version"))
	currentVersion, err := s.querier.GetUserTokenVersion(dbCtx, payload.UserID)
	dbSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("get token version: %w", err)
	}

	// 3. Version mismatch = token invalidated (password changed, security event)
	if payload.Version != int(*currentVersion) {
		tracing.AddEvent(span, "token_version_mismatch",
			tracing.Int("token.version", payload.Version),
			tracing.Int("current.version", int(*currentVersion)),
		)
		return nil, token.ErrInvalidToken
	}

	tracing.AddEvent(span, "token_verified")
	tracing.SetOK(span)
	return payload, nil
}

// RefreshAccessToken exchanges a refresh token for a new access token
func (s *TokenService) RefreshAccessToken(ctx context.Context, refreshTokenString string) (*model.TokenPair, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "TokenService", "RefreshAccessToken")
	defer span.End()

	// 1. Verify refresh token JWT signature and expiration
	tracing.AddEvent(span, "verifying_refresh_token")
	payload, err := s.tokenMaker.VerifyToken(refreshTokenString)
	if err != nil {
		tracing.AddEvent(span, "refresh_token_invalid")
		tracing.RecordError(span, err)
		return nil, token.ErrInvalidToken
	}

	tracing.SetAttributes(span,
		tracing.String("user.id", payload.UserID.String()),
		tracing.String("token.id", payload.ID.String()),
	)

	// 2. Hash the token to lookup in database
	tokenHash := hashToken(refreshTokenString)

	// 3. Get refresh token from database (including revoked ones for security check)
	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "SELECT", "refresh_tokens")
	storedToken, err := s.querier.GetRefreshTokenIncludingRevoked(dbCtx, tokenHash)
	dbSpan.End()

	if err != nil {
		tracing.AddEvent(span, "refresh_token_not_found")
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}

	// 4. SECURITY CHECK: Detect token reuse (revoked token being used)
	if storedToken.IsRevoked {
		// 🚨 CRITICAL SECURITY EVENT!
		tracing.AddEvent(span, "security_alert_revoked_token_reuse",
			tracing.String("user.id", storedToken.UserID.String()),
			tracing.String("token.id", storedToken.ID.String()),
		)

		// Log the security incident
		fmt.Printf("🚨 SECURITY ALERT: Revoked token reuse detected!\n")
		fmt.Printf("   User ID: %s\n", storedToken.UserID)
		fmt.Printf("   Token ID: %s\n", storedToken.ID)
		fmt.Printf("   Revoked At: %s\n", storedToken.LastUsedAt.Time)
		fmt.Printf("   Time Since Revoke: %s\n", time.Since(storedToken.LastUsedAt.Time))

		// Security Response: Revoke ALL user tokens immediately
		revokeErr := s.RevokeAllUserTokens(ctx, storedToken.UserID)
		if revokeErr != nil {
			tracing.AddEvent(span, "failed_to_revoke_all_tokens", tracing.String("error", revokeErr.Error()))
			fmt.Printf("Failed to revoke all tokens: %v\n", revokeErr)
		}

		tracing.RecordError(span, ErrSecurityAlert)
		return nil, ErrSecurityAlert
	}

	// 5. Check if token is expired
	if time.Now().After(storedToken.ExpiresAt.Time) {
		tracing.AddEvent(span, "refresh_token_expired")
		return nil, token.ErrExpiredToken
	}

	// 6. Update last used timestamp (for session tracking)
	updateCtx, updateSpan := tracing.StartDBSpan(ctx, "UPDATE", "refresh_tokens")
	_ = s.querier.UpdateRefreshTokenUsage(updateCtx, storedToken.ID)
	updateSpan.End()

	// 7. Get current user data (fresh from DB)
	userCtx, userSpan := tracing.StartDBSpan(ctx, "SELECT", "users")
	user, err := s.querier.GetUserByID(userCtx, payload.UserID)
	userSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 8. Check if user is still active (not banned/deleted)
	// TODO: Add user status check if you have user.IsActive field
	// if !user.IsActive {
	//     return nil, errors.New("user account disabled")
	// }

	// 9. Create new access token with current user data and version
	tracing.AddEvent(span, "creating_new_access_token")
	accessToken, _, err := s.tokenMaker.CreateToken(token.TokenParams{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     string(user.Role),
		Version:  int(*user.TokenVersion),
		Duration: 15 * time.Minute,
	})
	if err != nil {
		tracing.RecordError(span, err)
		return nil, fmt.Errorf("create access token: %w", err)
	}

	tracing.AddEvent(span, "access_token_refreshed")
	tracing.SetOK(span)
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
	ctx, span := tracing.StartServiceSpan(ctx, "TokenService", "Logout")
	defer span.End()

	tokenHash := hashToken(refreshTokenString)

	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "UPDATE", "refresh_tokens")
	tracing.SetAttributes(dbSpan, tracing.String("db.query_type", "revoke_token"))
	err := s.querier.RevokeRefreshToken(dbCtx, tokenHash)
	dbSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return err
	}

	tracing.AddEvent(span, "user_logged_out")
	tracing.SetOK(span)
	return nil
}

// RevokeAllUserTokens revokes all tokens for a user (logout all devices)
// Use when: password change, security breach, admin action
func (s *TokenService) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	ctx, span := tracing.StartServiceSpan(ctx, "TokenService", "RevokeAllUserTokens")
	defer span.End()

	tracing.SetAttributes(span, tracing.String("user.id", userID.String()))
	tracing.AddEvent(span, "revoking_all_user_tokens")

	// 1. Increment token version (invalidates ALL access tokens immediately)
	versionCtx, versionSpan := tracing.StartDBSpan(ctx, "UPDATE", "users")
	tracing.SetAttributes(versionSpan, tracing.String("db.query_type", "increment_token_version"))
	_, err := s.querier.IncrementUserTokenVersion(versionCtx, userID)
	versionSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("increment token version: %w", err)
	}

	// 2. Revoke all refresh tokens (logout all devices)
	revokeCtx, revokeSpan := tracing.StartDBSpan(ctx, "UPDATE", "refresh_tokens")
	tracing.SetAttributes(revokeSpan, tracing.String("db.query_type", "revoke_all_tokens"))
	err = s.querier.RevokeAllUserRefreshTokens(revokeCtx, userID)
	revokeSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}

	tracing.AddEvent(span, "all_tokens_revoked")
	tracing.SetOK(span)
	return nil
}

// GetUserSessions returns all active sessions for a user
func (s *TokenService) GetUserSessions(ctx context.Context, userID uuid.UUID, currentTokenID uuid.UUID) ([]model.Session, error) {
	ctx, span := tracing.StartServiceSpan(ctx, "TokenService", "GetUserSessions")
	defer span.End()

	tracing.SetAttributes(span,
		tracing.String("user.id", userID.String()),
		tracing.String("current_token.id", currentTokenID.String()),
	)

	dbCtx, dbSpan := tracing.StartDBSpan(ctx, "SELECT", "refresh_tokens")
	tracing.SetAttributes(dbSpan, tracing.String("db.query_type", "get_active_tokens"))
	tokens, err := s.querier.GetUserActiveRefreshTokens(dbCtx, userID)
	dbSpan.End()

	if err != nil {
		tracing.RecordError(span, err)
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

	tracing.SetAttributes(span, tracing.Int("session.count", len(sessions)))
	tracing.SetOK(span)
	return sessions, nil
}
