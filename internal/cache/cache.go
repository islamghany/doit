package cache

import (
	"context"
	"errors"
	"time"
)

// Error Handling
var (
	ErrCacheMiss     = errors.New("cache miss")                // Key not found (distinguish from actual errors)
	ErrExpired       = errors.New("cache expired")             // Key found but expired
	ErrConnection    = errors.New("cache connection error")    // Connection error
	ErrSerialization = errors.New("cache serialization error") // Marshal/unmarshal failed
	ErrInvalidKey    = errors.New("invalid cache key")         // Empty or invalid key
)

// Cache defines methods for caching operations.
type Cache interface {
	// Basic Operations
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// TTL Operations
	SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error // Set TTL to 0 to remove

	// Numbers Operations (useful for counters, rate limiting)
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	IncrementBy(ctx context.Context, key string, delta int64) (int64, error)

	// Utility Operations
	Clear(ctx context.Context) error
	Close() error

	// Health Check
	Ping(ctx context.Context) error
}

// validateKey checks if the key is valid
func validateKey(key string) error {
	if key == "" {
		return ErrInvalidKey
	}
	return nil
}
