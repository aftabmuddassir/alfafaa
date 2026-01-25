package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// RateLimiter is a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string]*clientInfo
	mu       sync.RWMutex
	limit    int
	duration time.Duration
}

type clientInfo struct {
	count     int
	resetTime time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, duration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    limit,
		duration: duration,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	info, exists := rl.requests[key]
	if !exists || now.After(info.resetTime) {
		rl.requests[key] = &clientInfo{
			count:     1,
			resetTime: now.Add(rl.duration),
		}
		return true
	}

	if info.count >= rl.limit {
		return false
	}

	info.count++
	return true
}

// GetRemaining returns the remaining requests for a key
func (rl *RateLimiter) GetRemaining(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info, exists := rl.requests[key]
	if !exists {
		return rl.limit
	}

	if time.Now().After(info.resetTime) {
		return rl.limit
	}

	return rl.limit - info.count
}

// cleanup periodically removes expired entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, info := range rl.requests {
			if now.After(info.resetTime) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use IP address as key, or user ID if authenticated
		key := c.ClientIP()
		if userID := GetUserID(c); userID != "" {
			key = "user:" + userID
		}

		if !limiter.Allow(key) {
			utils.ErrorResponseJSON(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests. Please try again later.", nil)
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Remaining", string(rune(limiter.GetRemaining(key))))

		c.Next()
	}
}

// RateLimitByIP creates a rate limiting middleware that uses IP address
func RateLimitByIP(limit int, duration time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, duration)
	return RateLimitMiddleware(limiter)
}
