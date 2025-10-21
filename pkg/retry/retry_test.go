package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestConnectWithRetry_Success(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	conn, err := ConnectWithRetry(context.Background(), cfg,
		func(ctx context.Context) (*mockConnection, error) {
			attempts++
			return &mockConnection{id: "conn-2"}, nil
		},
	)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if conn == nil {
		t.Error("expected connection, got nil")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got: %d", attempts)
	}
}

func TestConnectWithRetry_SuccessAfterRetries(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	conn, err := ConnectWithRetry(context.Background(), cfg,
		func(ctx context.Context) (*mockConnection, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("connection refused")
			}
			return &mockConnection{id: "conn-success"}, nil
		},
	)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if conn == nil {
		t.Error("expected connection, got nil")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got: %d", attempts)
	}
	if conn.id != "conn-success" {
		t.Errorf("expected conn-success, got: %s", conn.id)
	}
}

func TestConnectWithRetry_MaxAttemptsExceeded(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	expectedErr := errors.New("persistent error")

	conn, err := ConnectWithRetry(context.Background(), cfg,
		func(ctx context.Context) (*mockConnection, error) {
			attempts++
			return nil, expectedErr
		},
	)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if conn != nil {
		t.Error("expected nil connection, got:", conn)
	}
	if attempts != 4 { // Initial + 3 retries
		t.Errorf("expected 4 attempts, got: %d", attempts)
	}
	if !errors.Is(err, expectedErr) {
		t.Error("expected error to wrap original error")
	}
}

func TestConnectWithRetry_ContextCancellation(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	attempts := 0
	conn, err := ConnectWithRetry(ctx, cfg,
		func(ctx context.Context) (*mockConnection, error) {
			attempts++
			return nil, errors.New("error")
		},
	)

	if err == nil {
		t.Error("expected error due to context cancellation")
	}
	if conn != nil {
		t.Error("expected nil connection")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestConnectWithRetry_NoRetries(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  0, // No retries
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	conn, err := ConnectWithRetry(context.Background(), cfg,
		func(ctx context.Context) (*mockConnection, error) {
			attempts++
			if attempts == 1 {
				return nil, errors.New("first attempt fails")
			}
			return &mockConnection{id: "should-not-reach"}, nil
		},
	)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if conn != nil {
		t.Error("expected nil connection")
	}
	if attempts != 1 {
		t.Errorf("expected exactly 1 attempt (no retries), got: %d", attempts)
	}
}

func TestConnectWithRetry_OnRetryCallback(t *testing.T) {
	callbackCalls := 0
	var capturedAttempts []int
	var capturedDelays []time.Duration

	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
		OnRetry: func(attempt int, err error, delay time.Duration) {
			callbackCalls++
			capturedAttempts = append(capturedAttempts, attempt)
			capturedDelays = append(capturedDelays, delay)
		},
	}

	attempts := 0
	_, _ = ConnectWithRetry(context.Background(), cfg,
		func(ctx context.Context) (*mockConnection, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("error")
			}
			return &mockConnection{id: "success"}, nil
		},
	)

	if callbackCalls != 2 {
		t.Errorf("expected 2 callback calls, got: %d", callbackCalls)
	}
	if len(capturedAttempts) != 2 {
		t.Errorf("expected 2 captured attempts, got: %d", len(capturedAttempts))
	}
}

func TestExecuteWithRetry_Success(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	err := ExecuteWithRetry(context.Background(), cfg,
		func(ctx context.Context) error {
			attempts++
			if attempts < 2 {
				return errors.New("temporary error")
			}
			return nil
		},
	)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got: %d", attempts)
	}
}

func TestCalculateDelay_WithoutJitter(t *testing.T) {
	cfg := RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}

	testCases := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 10 * time.Second}, // Capped at MaxDelay
		{5, 10 * time.Second},
	}

	for _, tc := range testCases {
		actual := calculateDelay(tc.attempt, cfg)
		if actual != tc.expected {
			t.Errorf("attempt %d: expected %v, got %v", tc.attempt, tc.expected, actual)
		}
	}
}

func TestCalculateDelay_WithJitter(t *testing.T) {
	cfg := RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}

	// Test that jitter produces varying results
	delays := make(map[time.Duration]bool)
	for i := 0; i < 20; i++ {
		delay := calculateDelay(0, cfg)
		delays[delay] = true

		// Should be around 1 second (±20%)
		if delay < 800*time.Millisecond || delay > 1200*time.Millisecond {
			t.Errorf("delay %v out of expected range with jitter", delay)
		}
	}

	// With jitter, we should get variation
	if len(delays) == 1 {
		t.Error("expected variation in delays with jitter enabled")
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", cfg.MaxAttempts)
	}
	if cfg.InitialDelay != 1*time.Second {
		t.Errorf("expected InitialDelay=1s, got %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay=30s, got %v", cfg.MaxDelay)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", cfg.Multiplier)
	}
	if !cfg.Jitter {
		t.Error("expected Jitter=true")
	}
	if cfg.OnRetry == nil {
		t.Error("expected OnRetry to be set")
	}
}

// Mock connection for testing
type mockConnection struct {
	id string
}

func (m *mockConnection) Close() error {
	return nil
}

// Benchmark tests
func BenchmarkConnectWithRetry_Success(b *testing.B) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
		OnRetry:      nil, // Disable logging for benchmark
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConnectWithRetry(context.Background(), cfg,
			func(ctx context.Context) (*mockConnection, error) {
				return &mockConnection{id: "bench"}, nil
			},
		)
	}
}

func BenchmarkConnectWithRetry_WithRetries(b *testing.B) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
		OnRetry:      nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attempts := 0
		_, _ = ConnectWithRetry(context.Background(), cfg,
			func(ctx context.Context) (*mockConnection, error) {
				attempts++
				if attempts < 2 {
					return nil, errors.New("error")
				}
				return &mockConnection{id: "bench"}, nil
			},
		)
	}
}
