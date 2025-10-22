package database

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/tracelog"
)

// TestPgxLogger_Log tests the PgxLogger.Log method
func TestPgxLogger_Log(t *testing.T) {
	tests := []struct {
		name     string
		level    tracelog.LogLevel
		msg      string
		data     map[string]interface{}
		wantLog  bool
		contains []string
	}{
		{
			name:     "debug level message",
			level:    tracelog.LogLevelDebug,
			msg:      "debug message",
			data:     map[string]interface{}{"key": "value"},
			wantLog:  true,
			contains: []string{"PGX", "debug", "debug message", "key", "value"},
		},
		{
			name:     "info level message",
			level:    tracelog.LogLevelInfo,
			msg:      "info message",
			data:     map[string]interface{}{"operation": "query"},
			wantLog:  true,
			contains: []string{"PGX", "info", "info message", "operation"},
		},
		{
			name:     "error level message",
			level:    tracelog.LogLevelError,
			msg:      "error occurred",
			data:     map[string]interface{}{"error": "connection failed"},
			wantLog:  true,
			contains: []string{"PGX", "error", "error occurred", "connection failed"},
		},
		{
			name:     "message with empty data",
			level:    tracelog.LogLevelInfo,
			msg:      "simple message",
			data:     map[string]interface{}{},
			wantLog:  true,
			contains: []string{"PGX", "info", "simple message"},
		},
		{
			name:     "message with nil data",
			level:    tracelog.LogLevelWarn,
			msg:      "warning message",
			data:     nil,
			wantLog:  true,
			contains: []string{"PGX", "warn", "warning message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil) // Reset to default

			logger := NewPgxLogger()
			logger.Log(context.Background(), tt.level, tt.msg, tt.data)

			output := buf.String()

			if tt.wantLog && output == "" {
				t.Error("expected log output, got empty string")
			}

			// Check that expected strings are in the output
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected log output to contain '%s', got: %s", expected, output)
				}
			}
		})
	}
}

// TestNewPgxLogger tests that NewPgxLogger creates a valid logger
func TestNewPgxLogger(t *testing.T) {
	logger := NewPgxLogger()

	if logger == nil {
		t.Error("NewPgxLogger returned nil")
	}

	// Verify it implements the correct interface by calling Log
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	logger.Log(context.Background(), tracelog.LogLevelInfo, "test", nil)

	if buf.Len() == 0 {
		t.Error("logger.Log did not produce output")
	}
}

// TestPgxLogger_LogLevels tests all log levels
func TestPgxLogger_LogLevels(t *testing.T) {
	levels := []struct {
		level tracelog.LogLevel
		name  string
	}{
		{tracelog.LogLevelTrace, "trace"},
		{tracelog.LogLevelDebug, "debug"},
		{tracelog.LogLevelInfo, "info"},
		{tracelog.LogLevelWarn, "warn"},
		{tracelog.LogLevelError, "error"},
		{tracelog.LogLevelNone, "none"},
	}

	logger := NewPgxLogger()

	for _, l := range levels {
		t.Run(l.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			logger.Log(context.Background(), l.level, "test message", nil)

			output := buf.String()
			if !strings.Contains(output, l.name) {
				t.Errorf("expected log output to contain '%s', got: %s", l.name, output)
			}
		})
	}
}

// TestPgxLogger_WithContext tests logging with context values
func TestPgxLogger_WithContext(t *testing.T) {
	ctx := context.Background()
	// Add context values if needed
	// ctx = context.WithValue(ctx, "request_id", "123")

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	logger := NewPgxLogger()
	logger.Log(ctx, tracelog.LogLevelInfo, "message with context", map[string]interface{}{
		"user_id": 42,
		"action":  "query",
	})

	output := buf.String()

	if output == "" {
		t.Error("expected log output, got empty string")
	}

	// Verify data is included
	if !strings.Contains(output, "user_id") || !strings.Contains(output, "42") {
		t.Errorf("expected log to contain data fields, got: %s", output)
	}
}

// BenchmarkPgxLogger_Log benchmarks the logging performance
func BenchmarkPgxLogger_Log(b *testing.B) {
	logger := NewPgxLogger()
	ctx := context.Background()
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	// Discard output for benchmark
	log.SetOutput(bytes.NewBuffer(nil))
	defer log.SetOutput(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log(ctx, tracelog.LogLevelInfo, "benchmark message", data)
	}
}
