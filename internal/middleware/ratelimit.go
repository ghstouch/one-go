package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ghstouch/one-go/internal/model"
	"github.com/ghstouch/one-go/pkg/response"
)

// RateLimiter stores rate limit data
type RateLimiter struct {
	requests map[string]*requestData
	mu       sync.RWMutex
	limit    int
	window   time.Duration
	stop     chan struct{}
}

type requestData struct {
	count    int
	windowAt time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*requestData),
		limit:    limit,
		window:   window,
		stop:     make(chan struct{}),
	}

	go rl.cleanup()

	return rl
}

// Stop shuts down the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stop)
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		if userID := GetUserIDFromContext(c); userID != "" {
			key = "user:" + userID
		}

		if !limiter.Allow(key) {
			response.TooManyRequests(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow checks if request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	data, exists := rl.requests[key]
	if !exists || now.After(data.windowAt.Add(rl.window)) {
		rl.requests[key] = &requestData{
			count:    1,
			windowAt: now,
		}
		return true
	}

	if data.count >= rl.limit {
		return false
	}

	data.count++
	return true
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stop:
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for key, data := range rl.requests {
				if now.After(data.windowAt.Add(rl.window * 2)) {
					delete(rl.requests, key)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// GetRemainingRequests returns remaining requests for a key
func (rl *RateLimiter) GetRemainingRequests(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	data, exists := rl.requests[key]
	if !exists {
		return rl.limit
	}

	now := time.Now()
	if now.After(data.windowAt.Add(rl.window)) {
		return rl.limit
	}

	remaining := rl.limit - data.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// APIKeyRateLimiter enforces per-API-key rate limits
type APIKeyRateLimiter struct {
	perMinute map[string]*requestData
	perDay    map[string]*requestData
	mu        sync.RWMutex
	stop      chan struct{}
}

// NewAPIKeyRateLimiter creates a new per-key rate limiter
func NewAPIKeyRateLimiter() *APIKeyRateLimiter {
	rl := &APIKeyRateLimiter{
		perMinute: make(map[string]*requestData),
		perDay:    make(map[string]*requestData),
		stop:      make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// Stop shuts down the cleanup goroutine
func (rl *APIKeyRateLimiter) Stop() {
	close(rl.stop)
}

// Allow checks if a request is allowed for the given API key
func (rl *APIKeyRateLimiter) Allow(keyID string, maxPerMinute, maxPerDay int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Check per-minute limit
	if maxPerMinute > 0 {
		data, exists := rl.perMinute[keyID]
		if !exists || now.After(data.windowAt.Add(time.Minute)) {
			rl.perMinute[keyID] = &requestData{count: 1, windowAt: now}
		} else {
			if data.count >= maxPerMinute {
				return false
			}
			data.count++
		}
	}

	// Check per-day limit
	if maxPerDay > 0 {
		data, exists := rl.perDay[keyID]
		if !exists || now.After(data.windowAt.Add(24*time.Hour)) {
			rl.perDay[keyID] = &requestData{count: 1, windowAt: now}
		} else {
			if data.count >= maxPerDay {
				return false
			}
			data.count++
		}
	}

	return true
}

// cleanup removes old entries periodically
func (rl *APIKeyRateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stop:
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for key, data := range rl.perMinute {
				if now.After(data.windowAt.Add(2 * time.Minute)) {
					delete(rl.perMinute, key)
				}
			}
			for key, data := range rl.perDay {
				if now.After(data.windowAt.Add(25 * time.Hour)) {
					delete(rl.perDay, key)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// APIKeyRateLimitMiddleware enforces per-API-key rate limits
func APIKeyRateLimitMiddleware(limiter *APIKeyRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyVal, exists := c.Get(APIKeyContextKey)
		if !exists {
			c.Next()
			return
		}

		apiKey, ok := apiKeyVal.(*model.ApiKey)
		if !ok {
			c.Next()
			return
		}

		if !limiter.Allow(apiKey.ID, apiKey.MaxRequestsPerMinute, apiKey.MaxRequestsPerDay) {
			response.TooManyRequests(c, fmt.Sprintf("API key rate limit exceeded (max %d/min, %d/day)", apiKey.MaxRequestsPerMinute, apiKey.MaxRequestsPerDay))
			c.Abort()
			return
		}

		c.Next()
	}
}
