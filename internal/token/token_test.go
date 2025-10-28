package token

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPayload(t *testing.T) {
	tests := []struct {
		name   string
		params TokenParams
	}{
		{
			name: "creates payload with valid parameters",
			params: TokenParams{
				UserID:   uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
				Version:  1,
				Duration: time.Hour,
			},
		},
		{
			name: "creates payload with different duration",
			params: TokenParams{
				UserID:   uuid.New(),
				Email:    "admin@example.com",
				Username: "admin",
				Version:  2,
				Duration: 24 * time.Hour,
			},
		},
		{
			name: "creates payload with zero version",
			params: TokenParams{
				UserID:   uuid.New(),
				Email:    "user@example.com",
				Username: "user",
				Version:  0,
				Duration: 30 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeTime := time.Now()
			payload := NewPayload(tt.params)
			afterTime := time.Now()

			// Assert basic fields
			assert.NotEqual(t, uuid.Nil, payload.ID, "payload ID should not be nil")
			assert.Equal(t, tt.params.UserID, payload.UserID, "user ID should match")
			assert.Equal(t, tt.params.Email, payload.Email, "email should match")
			assert.Equal(t, tt.params.Username, payload.Username, "username should match")
			assert.Equal(t, tt.params.Version, payload.Version, "version should match")

			// Assert time fields are within reasonable bounds
			assert.True(t, payload.IssuedAt.After(beforeTime.Add(-time.Second)), "issued at should be after before time")
			assert.True(t, payload.IssuedAt.Before(afterTime.Add(time.Second)), "issued at should be before after time")

			// Assert expiration is correct (within reasonable bounds due to time precision)
			actualDuration := payload.ExpiredAt.Sub(payload.IssuedAt)
			assert.InDelta(t, tt.params.Duration.Nanoseconds(), actualDuration.Nanoseconds(), float64(time.Millisecond), "expired at should be issued at + duration")
		})
	}
}

func TestPayload_Valid(t *testing.T) {
	tests := []struct {
		name        string
		setupPayload func() *Payload
		wantErr     error
	}{
		{
			name: "valid payload - expires in future",
			setupPayload: func() *Payload {
				return &Payload{
					ID:        uuid.New(),
					UserID:    uuid.New(),
					Email:     "test@example.com",
					Username:  "testuser",
					Version:   1,
					IssuedAt:  time.Now(),
					ExpiredAt: time.Now().Add(time.Hour),
				}
			},
			wantErr: nil,
		},
		{
			name: "valid payload - expires in 1 second",
			setupPayload: func() *Payload {
				return &Payload{
					ID:        uuid.New(),
					UserID:    uuid.New(),
					Email:     "test@example.com",
					Username:  "testuser",
					Version:   1,
					IssuedAt:  time.Now(),
					ExpiredAt: time.Now().Add(time.Second),
				}
			},
			wantErr: nil,
		},
		{
			name: "invalid payload - expired 1 hour ago",
			setupPayload: func() *Payload {
				return &Payload{
					ID:        uuid.New(),
					UserID:    uuid.New(),
					Email:     "test@example.com",
					Username:  "testuser",
					Version:   1,
					IssuedAt:  time.Now().Add(-2 * time.Hour),
					ExpiredAt: time.Now().Add(-time.Hour),
				}
			},
			wantErr: ErrExpiredToken,
		},
		{
			name: "invalid payload - expired 1 second ago",
			setupPayload: func() *Payload {
				return &Payload{
					ID:        uuid.New(),
					UserID:    uuid.New(),
					Email:     "test@example.com",
					Username:  "testuser",
					Version:   1,
					IssuedAt:  time.Now().Add(-time.Minute),
					ExpiredAt: time.Now().Add(-time.Second),
				}
			},
			wantErr: ErrExpiredToken,
		},
		{
			name: "invalid payload - expires exactly now (edge case)",
			setupPayload: func() *Payload {
				now := time.Now()
				return &Payload{
					ID:        uuid.New(),
					UserID:    uuid.New(),
					Email:     "test@example.com",
					Username:  "testuser",
					Version:   1,
					IssuedAt:  now.Add(-time.Hour),
					ExpiredAt: now,
				}
			},
			// This might be valid or invalid depending on timing
			// The test allows for both scenarios
			wantErr: nil, // or ErrExpiredToken depending on exact timing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := tt.setupPayload()
			err := payload.Valid()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "error should match expected error")
			} else {
				// For edge cases near expiration, accept either result
				if tt.name == "invalid payload - expires exactly now (edge case)" {
					if err != nil {
						assert.ErrorIs(t, err, ErrExpiredToken)
					}
				} else {
					assert.NoError(t, err, "should not return error")
				}
			}
		})
	}
}

func TestPayload_UniqueIDs(t *testing.T) {
	// Create multiple payloads and ensure IDs are unique
	params := TokenParams{
		UserID:   uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Version:  1,
		Duration: time.Hour,
	}

	ids := make(map[uuid.UUID]bool)
	numPayloads := 100

	for i := 0; i < numPayloads; i++ {
		payload := NewPayload(params)
		require.NotEqual(t, uuid.Nil, payload.ID, "payload ID should not be nil")
		require.False(t, ids[payload.ID], "payload ID should be unique")
		ids[payload.ID] = true
	}

	assert.Equal(t, numPayloads, len(ids), "should have unique IDs for all payloads")
}

func TestPayload_DurationCalculation(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{"1 minute", time.Minute},
		{"15 minutes", 15 * time.Minute},
		{"1 hour", time.Hour},
		{"24 hours", 24 * time.Hour},
		{"7 days", 7 * 24 * time.Hour},
		{"30 days", 30 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := TokenParams{
				UserID:   uuid.New(),
				Email:    "test@example.com",
				Username: "testuser",
				Version:  1,
				Duration: tt.duration,
			}

			payload := NewPayload(params)

			actualDuration := payload.ExpiredAt.Sub(payload.IssuedAt)
			// Allow for small timing differences (within 1 millisecond)
			assert.InDelta(t, tt.duration.Nanoseconds(), actualDuration.Nanoseconds(), float64(time.Millisecond), "duration should be calculated correctly")

			// Verify token is valid immediately after creation
			assert.NoError(t, payload.Valid(), "newly created payload should be valid")
		})
	}
}

func TestPayload_ZeroDuration(t *testing.T) {
	params := TokenParams{
		UserID:   uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Version:  1,
		Duration: 0, // Zero duration
	}

	payload := NewPayload(params)

	// With zero duration, the token expires immediately
	// It may or may not be valid depending on exact timing
	err := payload.Valid()
	if err != nil {
		assert.ErrorIs(t, err, ErrExpiredToken, "zero duration token should be expired or about to expire")
	}
}

func TestPayload_NegativeDuration(t *testing.T) {
	params := TokenParams{
		UserID:   uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Version:  1,
		Duration: -time.Hour, // Negative duration
	}

	payload := NewPayload(params)

	// With negative duration, the token is already expired
	err := payload.Valid()
	assert.ErrorIs(t, err, ErrExpiredToken, "negative duration token should be expired")
	assert.True(t, payload.ExpiredAt.Before(payload.IssuedAt), "expired at should be before issued at")
}

