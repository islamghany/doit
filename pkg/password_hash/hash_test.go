package passwordHash

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password - standard length",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "valid password - minimum length",
			password: "pass",
			wantErr:  false,
		},
		{
			name:     "valid password - long password",
			password: "this_is_a_very_long_password_with_many_characters_1234567890",
			wantErr:  false,
		},
		{
			name:     "valid password - with special characters",
			password: "p@ssw0rd!#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "valid password - with unicode",
			password: "пароль密码🔐",
			wantErr:  false,
		},
		{
			name:     "valid password - with spaces",
			password: "my secure password",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords
		},
		{
			name:     "password with only spaces",
			password: "     ",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword([]byte(tt.password))

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hash)
				assert.NotEmpty(t, hash)

				// Hash should start with bcrypt prefix
				assert.True(t, strings.HasPrefix(string(hash), "$2a$") || 
					strings.HasPrefix(string(hash), "$2b$") ||
					strings.HasPrefix(string(hash), "$2y$"),
					"hash should have bcrypt prefix")

				// Hash should be at least 60 characters (bcrypt standard)
				assert.GreaterOrEqual(t, len(hash), 60, "bcrypt hash should be at least 60 bytes")
			}
		})
	}
}

func TestHashPassword_ProducesUniqueHashes(t *testing.T) {
	// Same password should produce different hashes (bcrypt uses random salt)
	password := []byte("same_password")

	hash1, err := HashPassword(password)
	require.NoError(t, err)

	hash2, err := HashPassword(password)
	require.NoError(t, err)

	hash3, err := HashPassword(password)
	require.NoError(t, err)

	// All hashes should be different due to random salt
	assert.NotEqual(t, hash1, hash2, "same password should produce different hashes")
	assert.NotEqual(t, hash2, hash3, "same password should produce different hashes")
	assert.NotEqual(t, hash1, hash3, "same password should produce different hashes")

	// But all should verify against the same password
	assert.NoError(t, ComparePassword(hash1, password))
	assert.NoError(t, ComparePassword(hash2, password))
	assert.NoError(t, ComparePassword(hash3, password))
}

func TestComparePassword(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		compareWith    string
		shouldMatch    bool
	}{
		{
			name:        "correct password",
			password:    "password123",
			compareWith: "password123",
			shouldMatch: true,
		},
		{
			name:        "wrong password",
			password:    "password123",
			compareWith: "wrongpassword",
			shouldMatch: false,
		},
		{
			name:        "case sensitive - uppercase",
			password:    "password123",
			compareWith: "PASSWORD123",
			shouldMatch: false,
		},
		{
			name:        "case sensitive - mixed case",
			password:    "PassWord123",
			compareWith: "password123",
			shouldMatch: false,
		},
		{
			name:        "extra spaces",
			password:    "password123",
			compareWith: "password123 ",
			shouldMatch: false,
		},
		{
			name:        "missing character",
			password:    "password123",
			compareWith: "password12",
			shouldMatch: false,
		},
		{
			name:        "extra character",
			password:    "password123",
			compareWith: "password1234",
			shouldMatch: false,
		},
		{
			name:        "empty password comparison",
			password:    "",
			compareWith: "",
			shouldMatch: true,
		},
		{
			name:        "empty vs non-empty",
			password:    "",
			compareWith: "password",
			shouldMatch: false,
		},
		{
			name:        "special characters match",
			password:    "p@ssw0rd!#$",
			compareWith: "p@ssw0rd!#$",
			shouldMatch: true,
		},
		{
			name:        "special characters mismatch",
			password:    "p@ssw0rd!#$",
			compareWith: "p@ssw0rd!#",
			shouldMatch: false,
		},
		{
			name:        "unicode password match",
			password:    "пароль密码🔐",
			compareWith: "пароль密码🔐",
			shouldMatch: true,
		},
		{
			name:        "unicode password mismatch",
			password:    "пароль密码🔐",
			compareWith: "пароль密码",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash the original password
			hash, err := HashPassword([]byte(tt.password))
			require.NoError(t, err)

			// Compare with the comparison password
			err = ComparePassword(hash, []byte(tt.compareWith))

			if tt.shouldMatch {
				assert.NoError(t, err, "passwords should match")
			} else {
				assert.Error(t, err, "passwords should not match")
				assert.ErrorIs(t, err, bcrypt.ErrMismatchedHashAndPassword)
			}
		})
	}
}

func TestComparePassword_WithInvalidHash(t *testing.T) {
	tests := []struct {
		name     string
		hash     []byte
		password []byte
	}{
		{
			name:     "empty hash",
			hash:     []byte(""),
			password: []byte("password"),
		},
		{
			name:     "invalid hash format",
			hash:     []byte("not_a_valid_bcrypt_hash"),
			password: []byte("password"),
		},
		{
			name:     "truncated hash",
			hash:     []byte("$2a$10$abc123"),
			password: []byte("password"),
		},
		{
			name:     "random bytes",
			hash:     []byte{0x00, 0x01, 0x02, 0x03},
			password: []byte("password"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ComparePassword(tt.hash, tt.password)
			assert.Error(t, err, "should error with invalid hash")
		})
	}
}

func TestHashPassword_Concurrent(t *testing.T) {
	// Test that hashing is safe for concurrent use
	password := []byte("concurrent_password")
	numGoroutines := 100
	done := make(chan []byte, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			hash, err := HashPassword(password)
			require.NoError(t, err)
			done <- hash
		}()
	}

	// Collect all hashes
	hashes := make([][]byte, 0, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		hash := <-done
		hashes = append(hashes, hash)
		
		// Each hash should be valid
		err := ComparePassword(hash, password)
		assert.NoError(t, err)
	}

	// All hashes should be unique (due to random salt)
	uniqueHashes := make(map[string]bool)
	for _, hash := range hashes {
		uniqueHashes[string(hash)] = true
	}
	assert.Equal(t, numGoroutines, len(uniqueHashes), "all hashes should be unique")
}

func TestComparePassword_Concurrent(t *testing.T) {
	// Test that comparison is safe for concurrent use
	password := []byte("concurrent_password")
	hash, err := HashPassword(password)
	require.NoError(t, err)

	numGoroutines := 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			err := ComparePassword(hash, password)
			done <- err == nil
		}()
	}

	// All comparisons should succeed
	for i := 0; i < numGoroutines; i++ {
		result := <-done
		assert.True(t, result)
	}
}

func TestBcryptCostVerification(t *testing.T) {
	// Verify that we're using the expected cost factor
	password := []byte("test_password")
	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Extract cost from hash (bcrypt format: $2a$10$...)
	cost, err := bcrypt.Cost(hash)
	require.NoError(t, err)
	
	// Should use DefaultCost (10)
	assert.Equal(t, bcrypt.DefaultCost, cost, "should use bcrypt.DefaultCost")
}

// Benchmark tests
func BenchmarkHashPassword(b *testing.B) {
	password := []byte("benchmark_password")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

func BenchmarkComparePassword(b *testing.B) {
	password := []byte("benchmark_password")
	hash, _ := HashPassword(password)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComparePassword(hash, password)
	}
}

func BenchmarkHashPassword_DifferentLengths(b *testing.B) {
	lengths := []int{8, 16, 32, 64, 72, 100}
	
	for _, length := range lengths {
		password := []byte(strings.Repeat("a", length))
		b.Run(fmt.Sprintf("length_%d", length), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = HashPassword(password)
			}
		})
	}
}