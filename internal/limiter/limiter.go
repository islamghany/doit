package limiter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"doit/internal/cache"
	"doit/internal/model"
	"doit/internal/web"
	"doit/pkg/logger"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// RateLimiter handles rate limiting using cache backend
type RateLimiter struct {
	cache  cache.Cache
	logger *logger.Logger
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cache cache.Cache, logger *logger.Logger) *RateLimiter {
	return &RateLimiter{
		cache:  cache,
		logger: logger,
	}
}

// RateLimitResult contains rate limit check results
type RateLimitResult struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetAt    time.Time
	RetryAfter time.Duration
}

// tokenBucket represents the state of a token bucket
type tokenBucket struct {
	Tokens     float64   `json:"tokens"`      // Current tokens in bucket
	LastRefill time.Time `json:"last_refill"` // Last time bucket was refilled
	Capacity   int       `json:"capacity"`    // Maximum bucket capacity
	RefillRate float64   `json:"refill_rate"` // Tokens per second
}

// CheckLimit checks if request is within rate limit using sliding window algorithm
func (rl *RateLimiter) CheckLimit(ctx context.Context, identifier string, config RateLimitConfig) (*RateLimitResult, error) {
	key := fmt.Sprintf("%s:%s", config.KeyPrefix, identifier)

	// Get or create bucket
	bucket, err := rl.getBucket(ctx, key, config)
	if err != nil {
		// Cache error - fail open (allow request but log)
		rl.logger.Error(ctx, "rate limit cache error", "error", err, "identifier", identifier)
		return &RateLimitResult{
			Allowed:   true,
			Limit:     config.Limit,
			Remaining: config.Limit,
			ResetAt:   time.Now().Add(config.Window),
		}, nil
	}

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(bucket.LastRefill).Seconds()
	tokensToAdd := elapsed * bucket.RefillRate

	// Add tokens, but don't exceed capacity
	bucket.Tokens = math.Min(float64(bucket.Capacity), bucket.Tokens+tokensToAdd)
	bucket.LastRefill = now

	// Check if we have at least 1 token
	allowed := bucket.Tokens >= 1.0
	remaining := int(math.Floor(bucket.Tokens))

	if allowed {
		// Consume 1 token
		bucket.Tokens -= 1.0
		remaining = int(math.Floor(bucket.Tokens))
	} else {
		remaining = 0
	}

	// Calculate when next token will be available
	tokensNeeded := 1.0 - bucket.Tokens
	if tokensNeeded < 0 {
		tokensNeeded = 0
	}
	secondsUntilNextToken := tokensNeeded / bucket.RefillRate
	retryAfter := time.Duration(secondsUntilNextToken * float64(time.Second))

	// Calculate when bucket will be full
	tokensToFull := float64(bucket.Capacity) - bucket.Tokens
	secondsToFull := tokensToFull / bucket.RefillRate
	resetAt := now.Add(time.Duration(secondsToFull * float64(time.Second)))

	result := &RateLimitResult{
		Allowed:    allowed,
		Limit:      config.Limit,
		Remaining:  remaining,
		ResetAt:    resetAt,
		RetryAfter: retryAfter,
	}

	// Save bucket state
	if err := rl.saveBucket(ctx, key, bucket, config.Window); err != nil {
		// Log but don't fail the request
		rl.logger.Error(ctx, "failed to save rate limit bucket", "error", err)
	}

	// Log if rate limit exceeded
	if !allowed {
		rl.logger.Warn(ctx, "rate limit exceeded",
			"identifier", identifier,
			"prefix", config.KeyPrefix,
			"limit", config.Limit,
			"tokens_remaining", bucket.Tokens,
		)
	}

	return result, nil
}

// saveBucket saves the token bucket to cache
func (rl *RateLimiter) saveBucket(ctx context.Context, key string, bucket *tokenBucket, ttl time.Duration) error {
	// Serialize bucket
	data, err := json.Marshal(bucket)
	if err != nil {
		return fmt.Errorf("failed to marshal bucket: %w", err)
	}

	// Save with TTL (add buffer for clock skew)
	return rl.cache.SetWithTTL(ctx, key, string(data), ttl*2)
}

// getBucket retrieves or creates a token bucket
func (rl *RateLimiter) getBucket(ctx context.Context, key string, config RateLimitConfig) (*tokenBucket, error) {
	val, err := rl.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, cache.ErrCacheMiss) {
			// Create new bucket (starts full)
			return &tokenBucket{
				Tokens:     float64(config.Limit),
				LastRefill: time.Now(),
				Capacity:   config.Limit,
				RefillRate: config.RefillRate(),
			}, nil
		}
		return nil, err
	}

	// Deserialize bucket
	var bucket tokenBucket

	// Handle different types from cache
	switch v := val.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &bucket); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bucket: %w", err)
		}
	case map[string]interface{}:
		// JSON deserialization returns map[string]interface{}
		bucket = tokenBucket{
			Tokens:     v["tokens"].(float64),
			LastRefill: parseTime(v["last_refill"]),
			Capacity:   int(v["capacity"].(float64)),
			RefillRate: v["refill_rate"].(float64),
		}
	default:
		// Try to unmarshal as JSON
		jsonBytes, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize bucket value: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &bucket); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bucket: %w", err)
		}
	}

	return &bucket, nil
}

// parseTime handles time parsing from cache
func parseTime(val interface{}) time.Time {
	switch v := val.(type) {
	case string:
		t, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return time.Now()
		}
		return t
	case float64:
		// Unix timestamp
		return time.Unix(int64(v), 0)
	default:
		return time.Now()
	}
}

// IsWhitelisted checks if identifier is whitelisted
func (rl *RateLimiter) IsWhitelisted(identifier string, config RateLimitConfig) bool {
	if config.Whitelist == nil {
		return false
	}
	return config.Whitelist[identifier]
}

// GetIdentifier extracts identifier based on strategy
func (rl *RateLimiter) GetIdentifier(r *http.Request, strategy RateLimitStrategy) string {
	switch strategy {
	case StrategyIP:
		return web.GetClientIP(r)

	case StrategyUser:
		user := model.GetUserContext(r.Context())
		if user != nil {
			return user.ID.String()
		}
		// Fallback to IP if not authenticated
		return web.GetClientIP(r)

	case StrategyIPAndUser:
		ip := web.GetClientIP(r)
		user := model.GetUserContext(r.Context())
		if user != nil {
			return fmt.Sprintf("%s:%s", ip, user.ID.String())
		}
		return ip

	case StrategyGlobal:
		return "global"

	default:
		return web.GetClientIP(r)
	}
}
