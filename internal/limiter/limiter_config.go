package limiter

import "time"

// Token Bucket Algorithm:
// ✅ Simple and intuitive
// ✅ Allows burst traffic naturally
// ✅ Smooth refill over time
// ✅ Industry standard (AWS, Stripe, etc. use this)
//
// How it works:
// - Each user has a bucket with capacity = Limit
// - Bucket starts full
// - Each request consumes 1 token
// - Tokens refill at constant rate: Limit tokens per Window
// - If bucket empty, request denied
//
// Example: Limit=5, Window=15min
// - Refill rate: 5 tokens / 15min = 1 token every 3 minutes
// - Allows burst of 5 requests immediately
// - Then 1 request every 3 minutes

// RateLimitStrategy defines how to identify rate limit subjects
type RateLimitStrategy string

const (
	// StrategyIP limits by client IP address
	StrategyIP RateLimitStrategy = "ip"

	// StrategyUser limits by authenticated user ID
	StrategyUser RateLimitStrategy = "user"

	// StrategyGlobal limits globally (all requests)
	StrategyGlobal RateLimitStrategy = "global"

	// StrategyIPAndUser combines IP and user (stricter)
	StrategyIPAndUser RateLimitStrategy = "ip_and_user"
)

// RateLimitConfig defines rate limiting parameters
type RateLimitConfig struct {
	// Strategy determines how to identify the subject
	Strategy RateLimitStrategy

	// Limit is the bucket capacity (maximum tokens)
	Limit int

	// Window is the time to refill all tokens (refill rate = Limit / Window)
	Window time.Duration

	// KeyPrefix is used to namespace cache keys
	KeyPrefix string

	// SkipSuccessful if true, only count failed requests (useful for login)
	SkipSuccessful bool

	// Whitelist contains IPs or user IDs that bypass rate limiting
	Whitelist map[string]bool
}

// RefillRate returns tokens per second
func (c RateLimitConfig) RefillRate() float64 {
	return float64(c.Limit) / c.Window.Seconds()
}

// EndpointLimits defines rate limits for specific endpoints
type EndpointLimits struct {
	// Login endpoints - strict limits
	Login RateLimitConfig

	// Register endpoint - prevent spam
	Register RateLimitConfig

	// Token refresh - normal usage
	Refresh RateLimitConfig

	// General API - fair usage
	General RateLimitConfig

	// Search/expensive operations
	Search RateLimitConfig
}

// DefaultEndpointLimits returns your specified rate limits
func DefaultEndpointLimits() *EndpointLimits {
	return &EndpointLimits{
		Login: RateLimitConfig{
			Strategy:       StrategyIP,
			Limit:          5,
			Window:         15 * time.Minute,
			KeyPrefix:      "rl:login",
			SkipSuccessful: false, // Count all attempts
			Whitelist:      make(map[string]bool),
		},
		Register: RateLimitConfig{
			Strategy:  StrategyIP,
			Limit:     3,
			Window:    60 * time.Minute,
			KeyPrefix: "rl:register",
			Whitelist: make(map[string]bool),
		},
		Refresh: RateLimitConfig{
			Strategy:  StrategyUser,
			Limit:     10,
			Window:    5 * time.Minute,
			KeyPrefix: "rl:refresh",
			Whitelist: make(map[string]bool),
		},
		General: RateLimitConfig{
			Strategy:  StrategyIP,
			Limit:     10,
			Window:    1 * time.Minute,
			KeyPrefix: "rl:api",
			Whitelist: make(map[string]bool),
		},
		Search: RateLimitConfig{
			Strategy:  StrategyUser,
			Limit:     30,
			Window:    1 * time.Minute,
			KeyPrefix: "rl:search",
			Whitelist: make(map[string]bool),
		},
	}
}
