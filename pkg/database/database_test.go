package database

import (
	"strings"
	"testing"
	"time"
)

// TestBuildDSN tests the DSN string construction
func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected map[string]string // key-value pairs that should be in DSN
	}{
		{
			name: "basic config with TLS disabled",
			config: Config{
				Host:       "localhost",
				Port:       5432,
				User:       "testuser",
				Password:   "testpass",
				Database:   "testdb",
				DisableTLS: true,
				MaxConns:   25,
			},
			expected: map[string]string{
				"host":           "localhost",
				"port":           "5432",
				"user":           "testuser",
				"password":       "testpass",
				"dbname":         "testdb",
				"sslmode":        "disable",
				"pool_max_conns": "25",
			},
		},
		{
			name: "config with TLS enabled",
			config: Config{
				Host:       "db.example.com",
				Port:       5433,
				User:       "admin",
				Password:   "securepass",
				Database:   "proddb",
				DisableTLS: false,
				MaxConns:   50,
			},
			expected: map[string]string{
				"host":           "db.example.com",
				"port":           "5433",
				"user":           "admin",
				"password":       "securepass",
				"dbname":         "proddb",
				"sslmode":        "require",
				"pool_max_conns": "50",
			},
		},
		{
			name: "config with special characters in password",
			config: Config{
				Host:       "localhost",
				Port:       5432,
				User:       "user",
				Password:   "p@ss!word#123",
				Database:   "db",
				DisableTLS: true,
				MaxConns:   10,
			},
			expected: map[string]string{
				"host":     "localhost",
				"password": "p@ss!word#123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := BuildDSN(tt.config)

			// Check that DSN is not empty
			if dsn == "" {
				t.Error("BuildDSN returned empty string")
			}

			// Verify expected key-value pairs are in the DSN
			for key, value := range tt.expected {
				expected := key + "=" + value
				if !strings.Contains(dsn, expected) {
					t.Errorf("DSN missing expected '%s', got: %s", expected, dsn)
				}
			}
		})
	}
}

// TestBuildDSN_SSLMode specifically tests SSL mode logic
func TestBuildDSN_SSLMode(t *testing.T) {
	tests := []struct {
		name            string
		disableTLS      bool
		expectedSSLMode string
	}{
		{
			name:            "TLS enabled",
			disableTLS:      false,
			expectedSSLMode: "sslmode=require",
		},
		{
			name:            "TLS disabled",
			disableTLS:      true,
			expectedSSLMode: "sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Host:       "localhost",
				Port:       5432,
				User:       "user",
				Password:   "pass",
				Database:   "db",
				DisableTLS: tt.disableTLS,
				MaxConns:   10,
			}

			dsn := BuildDSN(config)

			if !strings.Contains(dsn, tt.expectedSSLMode) {
				t.Errorf("expected DSN to contain '%s', got: %s", tt.expectedSSLMode, dsn)
			}
		})
	}
}

// TestBuildDSN_AllFieldsPresent ensures all config fields are used
func TestBuildDSN_AllFieldsPresent(t *testing.T) {
	config := Config{
		Host:       "testhost",
		Port:       5432,
		User:       "testuser",
		Password:   "testpass",
		Database:   "testdb",
		DisableTLS: true,
		MaxConns:   100,
	}

	dsn := BuildDSN(config)

	requiredFields := []string{
		"host=testhost",
		"port=5432",
		"user=testuser",
		"password=testpass",
		"dbname=testdb",
		"sslmode=",
		"pool_max_conns=100",
	}

	for _, field := range requiredFields {
		if !strings.Contains(dsn, field) {
			t.Errorf("DSN missing required field '%s', got: %s", field, dsn)
		}
	}
}

// TestConfig_DefaultValues tests that Config struct can hold values correctly
func TestConfig_DefaultValues(t *testing.T) {
	config := Config{
		Host:            "localhost",
		Port:            5432,
		Database:        "mydb",
		User:            "myuser",
		Password:        "mypass",
		MaxConns:        25,
		MinConns:        5,
		MaxConnLifetime: 1 * time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
		DisableTLS:      false,
		LogLevel:        "info",
	}

	// Verify all fields are set correctly
	if config.Host != "localhost" {
		t.Errorf("Host = %v, want localhost", config.Host)
	}
	if config.Port != 5432 {
		t.Errorf("Port = %v, want 5432", config.Port)
	}
	if config.MaxConns != 25 {
		t.Errorf("MaxConns = %v, want 25", config.MaxConns)
	}
	if config.MinConns != 5 {
		t.Errorf("MinConns = %v, want 5", config.MinConns)
	}
	if config.MaxConnLifetime != 1*time.Hour {
		t.Errorf("MaxConnLifetime = %v, want 1h", config.MaxConnLifetime)
	}
	if config.MaxConnIdleTime != 30*time.Minute {
		t.Errorf("MaxConnIdleTime = %v, want 30m", config.MaxConnIdleTime)
	}
	if config.DisableTLS != false {
		t.Errorf("DisableTLS = %v, want false", config.DisableTLS)
	}
	if config.LogLevel != "info" {
		t.Errorf("LogLevel = %v, want info", config.LogLevel)
	}
}

// BenchmarkBuildDSN benchmarks DSN construction
func BenchmarkBuildDSN(b *testing.B) {
	config := Config{
		Host:       "localhost",
		Port:       5432,
		User:       "user",
		Password:   "password",
		Database:   "db",
		DisableTLS: true,
		MaxConns:   25,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildDSN(config)
	}
}




