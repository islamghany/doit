package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts (0 means no retries)
	MaxAttempts int

	// InitialDelay is the delay before the first retry
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases after each retry
	Multiplier float64

	// Jitter adds randomness to prevent thundering herd
	// When true, adds up to 20% random variance to delays
	Jitter bool

	// OnRetry is called before each retry attempt (optional)
	OnRetry func(attempt int, err error, delay time.Duration)
}

// DefaultRetryConfig returns sensible defaults for connection retries
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		OnRetry: func(attempt int, err error, delay time.Duration) {
			fmt.Printf("⚠️  Connection attempt %d failed: %v. Retrying in %v...\n", attempt, err, delay)
		},
	}
}

// ConnectWithRetry is a generic function that retries a connection function
// T is the connection type (can be *pgxpool.Pool, *redis.Client, *http.Client, etc.)
// connectFn is your connection creation function that returns (connection, error)
//
// Example usage:
//
//	pool, err := ConnectWithRetry(ctx, cfg, func(ctx context.Context) (*pgxpool.Pool, error) {
//	    return pgxpool.New(ctx, connString)
//	})
func ConnectWithRetry[T any](ctx context.Context, cfg RetryConfig, connectFn func(context.Context) (T, error)) (T, error) {
	var connection T
	var lastErr error

	maxAttempts := cfg.MaxAttempts
	if maxAttempts < 0 {
		maxAttempts = 0
	}

	// Total attempts = initial attempt + retries
	totalAttempts := maxAttempts + 1

	for attempt := 0; attempt < totalAttempts; attempt++ {
		// Check if context is cancelled before attempting
		select {
		case <-ctx.Done():
			return connection, fmt.Errorf("connection cancelled: %w", ctx.Err())
		default:
		}

		// Attempt to create connection
		conn, err := connectFn(ctx)
		if err == nil {
			// Success!
			return conn, nil
		}

		// Store the error
		lastErr = err

		// If this is the last attempt, don't retry
		if attempt >= maxAttempts {
			break
		}

		// Calculate delay for next retry
		delay := calculateDelay(attempt, cfg)

		// Call retry callback if provided
		if cfg.OnRetry != nil {
			cfg.OnRetry(attempt+1, lastErr, delay)
		}

		// Wait before next retry with context cancellation support
		select {
		case <-ctx.Done():
			return connection, fmt.Errorf("connection cancelled after %d attempts: %w", attempt+1, ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return connection, fmt.Errorf("failed to connect after %d attempts: %w", totalAttempts, lastErr)
}

// ExecuteWithRetry retries a generic operation (doesn't return a value, only error)
// Useful for operations like migrations, seeds, or other idempotent operations
//
// Example usage:
//
//	err := ExecuteWithRetry(ctx, cfg, func(ctx context.Context) error {
//	    return runMigrations(db)
//	})
func ExecuteWithRetry(ctx context.Context, cfg RetryConfig, operation func(context.Context) error) error {
	_, err := ConnectWithRetry(ctx, cfg, func(ctx context.Context) (struct{}, error) {
		err := operation(ctx)
		return struct{}{}, err
	})
	return err
}

// calculateDelay computes the delay for a given retry attempt with exponential backoff
func calculateDelay(attempt int, cfg RetryConfig) time.Duration {
	// Calculate exponential delay: initialDelay * multiplier^attempt
	delayFloat := float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt))
	delay := time.Duration(delayFloat)

	// Cap at max delay
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}

	// Add jitter if enabled (±20% random variance)
	if cfg.Jitter {
		jitterRange := float64(delay) * 0.2                 // 20% of delay
		jitterDelta := (rand.Float64()*2 - 1) * jitterRange // Random value between -20% and +20%
		delay = time.Duration(float64(delay) + jitterDelta)

		// Ensure delay doesn't go negative or exceed max
		if delay < 0 {
			delay = cfg.InitialDelay
		}
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}

	return delay
}
