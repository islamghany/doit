package token

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	minSecretKeySize = 32
)

var (
	ErrInvalidSecretKey     = errors.New("secret key must be at least 32 characters")
	ErrInvalidSigningMethod = errors.New("invalid token signing method")
)

// JWTToken is a JWT implementation of TokenService.
type JWTToken struct {
	secretKey string
}

// CustomClaims represents the structured JWT claims.
type CustomClaims struct {
	ID       string `json:"id"`       // JTI - JWT ID
	UserID   string `json:"user_id"`  // Subject user ID
	Email    string `json:"email"`    // User email
	Username string `json:"username"` // Username
	Version  int    `json:"version"`  // Token version
	jwt.RegisteredClaims
}

// NewJWTToken creates a new JWTToken with validation.
func NewJWTToken(secretKey string) (*JWTToken, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, ErrInvalidSecretKey
	}

	return &JWTToken{
		secretKey: secretKey,
	}, nil
}

// CreateToken creates a new JWT token with structured claims.
func (t *JWTToken) CreateToken(params TokenParams) (string, *Payload, error) {
	payload := NewPayload(params)

	// Create structured claims
	claims := CustomClaims{
		ID:       payload.ID.String(),
		UserID:   payload.UserID.String(),
		Email:    payload.Email,
		Username: payload.Username,
		Version:  payload.Version,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        payload.ID.String(), // JTI
			Subject:   payload.UserID.String(),
			IssuedAt:  jwt.NewNumericDate(payload.IssuedAt),
			ExpiresAt: jwt.NewNumericDate(payload.ExpiredAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(t.secretKey))
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, payload, nil
}

// VerifyToken verifies and parses a JWT token with proper validation.
func (t *JWTToken) VerifyToken(tokenString string) (*Payload, error) {
	// Parse token with validation
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method to prevent algorithm confusion attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSigningMethod, token.Header["alg"])
		}
		return []byte(t.secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract and validate claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Parse UUIDs with proper error handling
	id, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid token ID: %w", err)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Construct payload
	payload := &Payload{
		ID:        id,
		UserID:    userID,
		Email:     claims.Email,
		Username:  claims.Username,
		Version:   claims.Version,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiredAt: claims.ExpiresAt.Time,
	}

	// Validate expiration
	if err := payload.Valid(); err != nil {
		return nil, err
	}

	return payload, nil
}
