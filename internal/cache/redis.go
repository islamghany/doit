package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"doit/internal/metrics"

	"github.com/redis/go-redis/v9"
)

// RedisOptions configures the Redis cache
type RedisOptions struct {
	// Addr is the Redis server address (host:port)
	Addr string

	// Password for Redis authentication (empty if no auth)
	Password string

	// DB is the Redis database number (0-15)
	DB int

	// PoolSize is the maximum number of socket connections
	PoolSize int

	// MinIdleConns is minimum number of idle connections
	MinIdleConns int

	// MaxRetries is the maximum number of retries before giving up
	MaxRetries int

	// DialTimeout is the timeout for establishing new connections
	DialTimeout time.Duration

	// ReadTimeout is the timeout for socket reads
	ReadTimeout time.Duration

	// WriteTimeout is the timeout for socket writes
	WriteTimeout time.Duration

	// DefaultTTL is the default time-to-live for cache entries (0 = no expiration)
	DefaultTTL time.Duration

	// Serializer handles value serialization (defaults to JSONSerializer)
	Serializer Serializer
}

// DefaultRedisOptions returns sensible defaults for Redis cache
func DefaultRedisOptions() *RedisOptions {
	return &RedisOptions{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		DefaultTTL:   0, // No expiration by default
		Serializer:   NewJSONSerializer(),
	}
}

// RedisCache implements the Cache interface
type RedisCache struct {
	client     *redis.Client
	options    *RedisOptions
	serializer Serializer
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(opts *RedisOptions) (*RedisCache, error) {
	if opts == nil {
		opts = DefaultRedisOptions()
	}

	if opts.Serializer == nil {
		opts.Serializer = NewJSONSerializer()
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         opts.Addr,
		Password:     opts.Password,
		DB:           opts.DB,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
		MaxRetries:   opts.MaxRetries,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnection, err)
	}

	return &RedisCache{
		client:     client,
		options:    opts,
		serializer: opts.Serializer,
	}, nil
}

// Set stores a value in Redis with optional default TTL
func (c *RedisCache) Set(ctx context.Context, key string, value any) error {
	return c.SetWithTTL(ctx, key, value, c.options.DefaultTTL)
}

// Get retrieves a value from Redis
func (c *RedisCache) Get(ctx context.Context, key string) (any, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			metrics.RecordCacheOperation("get", "redis", false) // miss
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("%w: %v", ErrConnection, err)
	}

	metrics.RecordCacheOperation("get", "redis", true) // hit
	// Deserialize
	var result any
	if err := c.serializer.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Delete removes a value from Redis
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrConnection, err)
	}

	return nil
}

// Exists checks if a key exists in Redis
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateKey(key); err != nil {
		return false, err
	}

	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrConnection, err)
	}

	return count > 0, nil
}

// SetWithTTL stores a value in Redis with specific TTL
func (c *RedisCache) SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error {
	if err := validateKey(key); err != nil {
		return err
	}

	// Serialize value
	data, err := c.serializer.Marshal(value)
	if err != nil {
		return err
	}

	// Set with TTL
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrConnection, err)
	}

	return nil
}

// GetTTL returns the remaining TTL for a key
func (c *RedisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if err := validateKey(key); err != nil {
		return 0, err
	}

	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrConnection, err)
	}

	// Redis returns -2 if key doesn't exist
	if ttl == -2*time.Second {
		return 0, ErrCacheMiss
	}

	// Redis returns -1 if key has no expiration
	if ttl == -1*time.Second {
		return -1, nil
	}

	return ttl, nil
}

// Expire sets a new TTL for a key
func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := validateKey(key); err != nil {
		return err
	}

	ok, err := c.client.Expire(ctx, key, ttl).Result()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnection, err)
	}

	if !ok {
		return ErrCacheMiss
	}

	return nil
}

// Increment increments a numeric value in Redis
func (c *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	if err := validateKey(key); err != nil {
		return 0, err
	}

	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrConnection, err)
	}

	return val, nil
}

// Decrement decrements a numeric value in Redis
func (c *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	if err := validateKey(key); err != nil {
		return 0, err
	}

	val, err := c.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrConnection, err)
	}

	return val, nil
}

// IncrementBy increments a value by delta (extra Redis-specific method)
func (c *RedisCache) IncrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	if err := validateKey(key); err != nil {
		return 0, err
	}

	val, err := c.client.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrConnection, err)
	}

	return val, nil
}

// Clear removes all keys from the current Redis database
func (c *RedisCache) Clear(ctx context.Context) error {
	if err := c.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrConnection, err)
	}
	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Ping checks if the Redis connection is alive
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
