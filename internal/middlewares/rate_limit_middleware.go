package middlewares

import (
	"fmt"
	"net/http"

	"doit/internal/limiter"
	"doit/internal/web"
)

var defaultEndpointLimits = limiter.DefaultEndpointLimits()

func RateLimitMiddleware(rateLimiter *limiter.RateLimiter, config limiter.RateLimitConfig) web.MiddleWare {
	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx := r.Context()

			// 1. Get identifier based on strategy
			identifier := rateLimiter.GetIdentifier(r, config.Strategy)

			// 2. Check whitelist
			if rateLimiter.IsWhitelisted(identifier, config) {
				return handler(w, r)
			}

			// 3.Check rate limit
			result, err := rateLimiter.CheckLimit(ctx, identifier, config)
			if err != nil {
				return handler(w, r)
			}

			// 4. Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))

			// 5. Return error if rate limit exceeded
			if !result.Allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(result.RetryAfter.Seconds())))
				return web.NewError(limiter.ErrRateLimitExceeded, http.StatusTooManyRequests)
			}

			// 6. Return response
			return handler(w, r)
		}
	}
}

// LoginRateLimitMiddleware applies strict rate limiting for login
func LoginRateLimitMiddleware(rateLimiter *limiter.RateLimiter) web.MiddleWare {
	config := defaultEndpointLimits.Login
	return RateLimitMiddleware(rateLimiter, config)
}

// RegisterRateLimitMiddleware applies rate limiting for registration
func RegisterRateLimitMiddleware(rateLimiter *limiter.RateLimiter) web.MiddleWare {
	config := defaultEndpointLimits.Register
	return RateLimitMiddleware(rateLimiter, config)
}

// RefreshRateLimitMiddleware applies rate limiting for token refresh
func RefreshRateLimitMiddleware(rateLimiter *limiter.RateLimiter) web.MiddleWare {
	config := defaultEndpointLimits.Refresh
	return RateLimitMiddleware(rateLimiter, config)
}

// GeneralRateLimitMiddleware applies rate limiting for general API usage
func GeneralRateLimitMiddleware(rateLimiter *limiter.RateLimiter) web.MiddleWare {
	config := defaultEndpointLimits.General
	return RateLimitMiddleware(rateLimiter, config)
}

// SearchRateLimitMiddleware applies rate limiting for search operations
func SearchRateLimitMiddleware(rateLimiter *limiter.RateLimiter) web.MiddleWare {
	config := defaultEndpointLimits.Search
	return RateLimitMiddleware(rateLimiter, config)
}
