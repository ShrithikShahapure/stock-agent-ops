package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	redisclient "github.com/shrithkshahapure/stock-agent-ops/internal/services/redis"
)

// RateLimiter creates rate limiting middleware
type RateLimiter struct {
	redis *redisclient.Client
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redis *redisclient.Client) *RateLimiter {
	return &RateLimiter{redis: redis}
}

// Limit returns middleware that rate limits requests
func (rl *RateLimiter) Limit(limit int, window time.Duration, keyPrefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if Redis is not available
			if rl.redis == nil || !rl.redis.IsConnected() {
				next.ServeHTTP(w, r)
				return
			}

			// Calculate window key
			windowSec := int64(window.Seconds())
			windowKey := redisclient.RateLimitKey(keyPrefix, time.Now().Unix()/windowSec)

			// Increment counter
			ctx := r.Context()
			count, err := rl.redis.Incr(ctx, windowKey)
			if err != nil {
				// On error, allow the request
				next.ServeHTTP(w, r)
				return
			}

			// Set expiry on first increment
			if count == 1 {
				rl.redis.Expire(ctx, windowKey, window)
			}

			// Check if over limit
			if count > int64(limit) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"detail": "Rate limit exceeded",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LimitWithTicker returns middleware that rate limits per ticker
func (rl *RateLimiter) LimitWithTicker(limit int, window time.Duration, keyPrefix string, tickerExtractor func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if Redis is not available
			if rl.redis == nil || !rl.redis.IsConnected() {
				next.ServeHTTP(w, r)
				return
			}

			// Build key with ticker if available
			key := keyPrefix
			if tickerExtractor != nil {
				if ticker := tickerExtractor(r); ticker != "" {
					key = keyPrefix + ":" + ticker
				}
			}

			// Calculate window key
			windowSec := int64(window.Seconds())
			windowKey := redisclient.RateLimitKey(key, time.Now().Unix()/windowSec)

			// Increment counter
			ctx := r.Context()
			count, err := rl.redis.Incr(ctx, windowKey)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Set expiry on first increment
			if count == 1 {
				rl.redis.Expire(ctx, windowKey, window)
			}

			// Check if over limit
			if count > int64(limit) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"detail": "Rate limit exceeded",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
