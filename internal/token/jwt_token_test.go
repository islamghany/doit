package token

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTToken(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		wantErr   error
	}{
		{
			name:      "valid secret key - exactly 32 characters",
			secretKey: "12345678901234567890123456789012",
			wantErr:   nil,
		},
		{
			name:      "valid secret key - more than 32 characters",
			secretKey: "this_is_a_very_long_secret_key_that_is_secure",
			wantErr:   nil,
		},
		{
			name:      "invalid secret key - empty",
			secretKey: "",
			wantErr:   ErrInvalidSecretKey,
		},
		{
			name:      "invalid secret key - too short (31 characters)",
			secretKey: "1234567890123456789012345678901",
			wantErr:   ErrInvalidSecretKey,
		},
		{
			name:      "invalid secret key - too short (16 characters)",
			secretKey: "1234567890123456",
			wantErr:   ErrInvalidSecretKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewJWTToken(tt.secretKey)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "error should match expected")
				assert.Nil(t, token, "token should be nil on error")
			} else {
				assert.NoError(t, err, "should not return error")
				assert.NotNil(t, token, "token should not be nil")
				assert.Equal(t, tt.secretKey, token.secretKey, "secret key should be stored")
			}
		})
	}
}

func TestJWTToken_CreateToken(t *testing.T) {
	secretKey := "test_secret_key_that_is_32_chars_long!!"
	jwtToken, err := NewJWTToken(secretKey)
	require.NoError(t, err)

	tests := []struct {
		name   string
		params TokenParams
	}{
		{
			name: "creates token with standard parameters",
			params: TokenParams{
				UserID:   uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
				Version:  1,
				Duration: time.Hour,
			},
		},
		{
			name: "creates token with different duration",
			params: TokenParams{
				UserID:   uuid.New(),
				Email:    "admin@example.com",
				Username: "admin",
				Version:  2,
				Duration: 24 * time.Hour,
			},
		},
		{
			name: "creates token with version 0",
			params: TokenParams{
				UserID:   uuid.New(),
				Email:    "user@example.com",
				Username: "user",
				Version:  0,
				Duration: 30 * time.Minute,
			},
		},
		{
			name: "creates token with special characters in email",
			params: TokenParams{
				UserID:   uuid.New(),
				Email:    "test+tag@example.co.uk",
				Username: "special_user",
				Version:  1,
				Duration: time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, payload, err := jwtToken.CreateToken(tt.params)

			require.NoError(t, err, "should not return error")
			assert.NotEmpty(t, tokenString, "token string should not be empty")
			assert.NotNil(t, payload, "payload should not be nil")

			// Verify payload fields
			assert.NotEqual(t, uuid.Nil, payload.ID, "payload ID should not be nil")
			assert.Equal(t, tt.params.UserID, payload.UserID, "user ID should match")
			assert.Equal(t, tt.params.Email, payload.Email, "email should match")
			assert.Equal(t, tt.params.Username, payload.Username, "username should match")
			assert.Equal(t, tt.params.Version, payload.Version, "version should match")

			// Verify token string format (JWT has 3 parts separated by dots)
			parts := strings.Split(tokenString, ".")
			assert.Equal(t, 3, len(parts), "JWT should have 3 parts")
		})
	}
}

func TestJWTToken_VerifyToken(t *testing.T) {
	secretKey := "test_secret_key_that_is_32_chars_long!!"
	jwtToken, err := NewJWTToken(secretKey)
	require.NoError(t, err)

	t.Run("successfully verifies valid token", func(t *testing.T) {
		params := TokenParams{
			UserID:   uuid.New(),
			Email:    "test@example.com",
			Username: "testuser",
			Version:  1,
			Duration: time.Hour,
		}

		tokenString, originalPayload, err := jwtToken.CreateToken(params)
		require.NoError(t, err)

		verifiedPayload, err := jwtToken.VerifyToken(tokenString)
		require.NoError(t, err)
		assert.NotNil(t, verifiedPayload, "verified payload should not be nil")

		// Compare all fields
		assert.Equal(t, originalPayload.ID, verifiedPayload.ID, "ID should match")
		assert.Equal(t, originalPayload.UserID, verifiedPayload.UserID, "user ID should match")
		assert.Equal(t, originalPayload.Email, verifiedPayload.Email, "email should match")
		assert.Equal(t, originalPayload.Username, verifiedPayload.Username, "username should match")
		assert.Equal(t, originalPayload.Version, verifiedPayload.Version, "version should match")
		// JWT uses Unix timestamps which only have second precision, so compare times in seconds
		assert.Equal(t, originalPayload.IssuedAt.Unix(), verifiedPayload.IssuedAt.Unix(), "issued at should match")
		assert.Equal(t, originalPayload.ExpiredAt.Unix(), verifiedPayload.ExpiredAt.Unix(), "expired at should match")
	})

	t.Run("fails to verify token with wrong secret", func(t *testing.T) {
		params := TokenParams{
			UserID:   uuid.New(),
			Email:    "test@example.com",
			Username: "testuser",
			Version:  1,
			Duration: time.Hour,
		}

		tokenString, _, err := jwtToken.CreateToken(params)
		require.NoError(t, err)

		// Create token service with different secret (must be at least 32 chars)
		wrongSecretToken, err := NewJWTToken("different_secret_key_must_be_32_characters_long!")
		require.NoError(t, err)

		payload, err := wrongSecretToken.VerifyToken(tokenString)
		assert.Error(t, err, "should return error")
		assert.Nil(t, payload, "payload should be nil")
	})

	t.Run("fails to verify expired token", func(t *testing.T) {
		params := TokenParams{
			UserID:   uuid.New(),
			Email:    "test@example.com",
			Username: "testuser",
			Version:  1,
			Duration: -time.Hour, // Negative duration = already expired
		}

		tokenString, _, err := jwtToken.CreateToken(params)
		require.NoError(t, err)

		payload, err := jwtToken.VerifyToken(tokenString)
		assert.Error(t, err, "should return error")
		// Error is wrapped, so check if it contains token expiration error
		assert.Contains(t, err.Error(), "token is expired", "should contain expired token error")
		assert.Nil(t, payload, "payload should be nil")
	})

	t.Run("fails to verify malformed token", func(t *testing.T) {
		tests := []struct {
			name  string
			token string
		}{
			{"empty token", ""},
			{"random string", "this_is_not_a_jwt_token"},
			{"incomplete JWT", "header.payload"},
			{"invalid base64", "not.valid.base64!!!"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload, err := jwtToken.VerifyToken(tt.token)
				assert.Error(t, err, "should return error")
				assert.Nil(t, payload, "payload should be nil")
			})
		}
	})

	t.Run("fails to verify token with invalid UUID", func(t *testing.T) {
		// Create a token with invalid UUID in claims
		claims := CustomClaims{
			ID:       "not-a-valid-uuid",
			UserID:   uuid.New().String(),
			Email:    "test@example.com",
			Username: "testuser",
			Version:  1,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        "not-a-valid-uuid",
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		require.NoError(t, err)

		payload, err := jwtToken.VerifyToken(tokenString)
		assert.Error(t, err, "should return error")
		assert.Nil(t, payload, "payload should be nil")
		assert.Contains(t, err.Error(), "invalid token ID", "error should mention invalid token ID")
	})
}

func TestJWTToken_AlgorithmValidation(t *testing.T) {
	secretKey := "test_secret_key_that_is_32_chars_long!!"
	jwtToken, err := NewJWTToken(secretKey)
	require.NoError(t, err)

	t.Run("rejects token with different algorithm", func(t *testing.T) {
		// Create a token with a different signing method (none)
		claims := CustomClaims{
			ID:       uuid.New().String(),
			UserID:   uuid.New().String(),
			Email:    "test@example.com",
			Username: "testuser",
			Version:  1,
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        uuid.New().String(),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}

		// Try to create a token with "none" algorithm (security vulnerability)
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		payload, err := jwtToken.VerifyToken(tokenString)
		assert.Error(t, err, "should reject token with different algorithm")
		assert.Nil(t, payload, "payload should be nil")
	})
}

func TestJWTToken_CreateAndVerifyRoundTrip(t *testing.T) {
	secretKey := "test_secret_key_that_is_32_chars_long!!"
	jwtToken, err := NewJWTToken(secretKey)
	require.NoError(t, err)

	// Test multiple tokens to ensure consistency
	numTokens := 10
	tokens := make(map[string]*Payload)

	// Create tokens
	for i := 0; i < numTokens; i++ {
		params := TokenParams{
			UserID:   uuid.New(),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Username: fmt.Sprintf("user%d", i),
			Version:  i,
			Duration: time.Hour,
		}

		tokenString, payload, err := jwtToken.CreateToken(params)
		require.NoError(t, err)
		tokens[tokenString] = payload
	}

	// Verify all tokens
	for tokenString, originalPayload := range tokens {
		verifiedPayload, err := jwtToken.VerifyToken(tokenString)
		require.NoError(t, err, "should verify token successfully")

		assert.Equal(t, originalPayload.ID, verifiedPayload.ID)
		assert.Equal(t, originalPayload.UserID, verifiedPayload.UserID)
		assert.Equal(t, originalPayload.Email, verifiedPayload.Email)
		assert.Equal(t, originalPayload.Username, verifiedPayload.Username)
		assert.Equal(t, originalPayload.Version, verifiedPayload.Version)
	}
}

func TestJWTToken_ConcurrentOperations(t *testing.T) {
	secretKey := "test_secret_key_that_is_32_chars_long!!"
	jwtToken, err := NewJWTToken(secretKey)
	require.NoError(t, err)

	// Test concurrent token creation and verification
	numGoroutines := 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			params := TokenParams{
				UserID:   uuid.New(),
				Email:    fmt.Sprintf("user%d@example.com", idx),
				Username: fmt.Sprintf("user%d", idx),
				Version:  idx,
				Duration: time.Hour,
			}

			// Create token
			tokenString, originalPayload, err := jwtToken.CreateToken(params)
			assert.NoError(t, err)

			// Verify token
			verifiedPayload, err := jwtToken.VerifyToken(tokenString)
			assert.NoError(t, err)
			assert.Equal(t, originalPayload.ID, verifiedPayload.ID)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestJWTToken_TokenExpiration(t *testing.T) {
	secretKey := "test_secret_key_that_is_32_chars_long!!"
	jwtToken, err := NewJWTToken(secretKey)
	require.NoError(t, err)

	t.Run("token expires after duration", func(t *testing.T) {
		params := TokenParams{
			UserID:   uuid.New(),
			Email:    "test@example.com",
			Username: "testuser",
			Version:  1,
			Duration: 2 * time.Second, // Short duration but not too short
		}

		tokenString, _, err := jwtToken.CreateToken(params)
		require.NoError(t, err)

		// Verify immediately - should succeed
		payload, err := jwtToken.VerifyToken(tokenString)
		assert.NoError(t, err)
		assert.NotNil(t, payload)

		// Wait for expiration (add buffer)
		time.Sleep(3 * time.Second)

		// Verify after expiration - should fail
		payload, err = jwtToken.VerifyToken(tokenString)
		assert.Error(t, err, "should return error after expiration")
		assert.Contains(t, err.Error(), "token is expired", "error should indicate token is expired")
		assert.Nil(t, payload)
	})
}

func TestJWTToken_RegisteredClaims(t *testing.T) {
	secretKey := "test_secret_key_that_is_32_chars_long!!"
	jwtToken, err := NewJWTToken(secretKey)
	require.NoError(t, err)

	params := TokenParams{
		UserID:   uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Version:  1,
		Duration: time.Hour,
	}

	tokenString, payload, err := jwtToken.CreateToken(params)
	require.NoError(t, err)

	// Parse token to check registered claims
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)

	claims, ok := token.Claims.(*CustomClaims)
	require.True(t, ok)

	// Verify registered claims
	assert.Equal(t, payload.ID.String(), claims.ID, "JTI should match payload ID")
	assert.Equal(t, payload.UserID.String(), claims.Subject, "Subject should match user ID")
	assert.NotNil(t, claims.IssuedAt, "IssuedAt should be set")
	assert.NotNil(t, claims.ExpiresAt, "ExpiresAt should be set")
}

func TestJWTToken_EmptyFields(t *testing.T) {
	secretKey := "test_secret_key_that_is_32_chars_long!!"
	jwtToken, err := NewJWTToken(secretKey)
	require.NoError(t, err)

	t.Run("handles empty email and username", func(t *testing.T) {
		params := TokenParams{
			UserID:   uuid.New(),
			Email:    "",
			Username: "",
			Version:  1,
			Duration: time.Hour,
		}

		tokenString, originalPayload, err := jwtToken.CreateToken(params)
		require.NoError(t, err)

		verifiedPayload, err := jwtToken.VerifyToken(tokenString)
		require.NoError(t, err)

		assert.Equal(t, originalPayload.Email, verifiedPayload.Email)
		assert.Equal(t, originalPayload.Username, verifiedPayload.Username)
		assert.Empty(t, verifiedPayload.Email)
		assert.Empty(t, verifiedPayload.Username)
	})
}

