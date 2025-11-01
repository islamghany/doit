package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

// TokenParams represents the parameters for creating a token.
type TokenParams struct {
	UserID   uuid.UUID
	Role     string
	Email    string
	Username string
	Version  int
	Duration time.Duration
}

// TokenMaker provides methods for managing tokens.
type TokenMaker interface {
	CreateToken(params TokenParams) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}

// Payload represents the payload of a token.
type Payload struct {
	ID        uuid.UUID `json:"id"`         // JTI - JWT ID for token identification/revocation
	UserID    uuid.UUID `json:"user_id"`    // Subject user ID
	Email     string    `json:"email"`      // User email
	Username  string    `json:"username"`   // Username
	Role      string    `json:"role"`       // User role
	Version   int       `json:"version"`    // Token version for invalidation
	IssuedAt  time.Time `json:"issued_at"`  // Token issue time
	ExpiredAt time.Time `json:"expired_at"` // Token expiration time
}

// NewPayload creates a new Payload.
func NewPayload(
	params TokenParams,
) *Payload {
	p := &Payload{
		ID:        uuid.New(),
		UserID:    params.UserID,
		Email:     params.Email,
		Username:  params.Username,
		Role:      params.Role,
		Version:   params.Version,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(params.Duration),
	}

	return p
}

// Valid checks if the token is valid or not.
func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiredAt) {
		return ErrExpiredToken
	}

	return nil
}
