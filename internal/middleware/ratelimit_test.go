package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	defer rl.Stop()

	if !rl.Allow("user1") {
		t.Fatal("first request should be allowed")
	}
	if !rl.Allow("user1") {
		t.Fatal("second request should be allowed")
	}
	if !rl.Allow("user1") {
		t.Fatal("third request should be allowed")
	}
	if rl.Allow("user1") {
		t.Fatal("fourth request should be blocked")
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	defer rl.Stop()

	if !rl.Allow("user1") {
		t.Fatal("user1 should be allowed")
	}
	if !rl.Allow("user2") {
		t.Fatal("user2 should be allowed")
	}
	if rl.Allow("user1") {
		t.Fatal("user1 should be blocked")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)
	defer rl.Stop()

	if !rl.Allow("user1") {
		t.Fatal("first request should be allowed")
	}
	if rl.Allow("user1") {
		t.Fatal("second request should be blocked")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("user1") {
		t.Fatal("request after window reset should be allowed")
	}
}

func TestRateLimiter_GetRemainingRequests(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	defer rl.Stop()

	if rl.GetRemainingRequests("user1") != 5 {
		t.Fatal("should have 5 remaining")
	}

	rl.Allow("user1")
	rl.Allow("user1")

	if rl.GetRemainingRequests("user1") != 3 {
		t.Fatalf("should have 3 remaining, got %d", rl.GetRemainingRequests("user1"))
	}
}

func TestAPIKeyRateLimiter_Allow(t *testing.T) {
	rl := NewAPIKeyRateLimiter()
	defer rl.Stop()

	// max 2 per minute, 10 per day
	if !rl.Allow("key1", 2, 10) {
		t.Fatal("first request should be allowed")
	}
	if !rl.Allow("key1", 2, 10) {
		t.Fatal("second request should be allowed")
	}
	if rl.Allow("key1", 2, 10) {
		t.Fatal("third request should be blocked (per-minute)")
	}
}

func TestAPIKeyRateLimiter_Unlimited(t *testing.T) {
	rl := NewAPIKeyRateLimiter()
	defer rl.Stop()

	// 0 = unlimited
	for i := 0; i < 100; i++ {
		if !rl.Allow("key1", 0, 0) {
			t.Fatalf("request %d should be allowed (unlimited)", i)
		}
	}
}

func TestAPIKeyRateLimiter_DifferentKeys(t *testing.T) {
	rl := NewAPIKeyRateLimiter()
	defer rl.Stop()

	rl.Allow("key1", 1, 100)
	rl.Allow("key2", 1, 100)

	// key1 should be blocked, key2 should also be blocked (used 1)
	if rl.Allow("key1", 1, 100) {
		t.Fatal("key1 should be blocked")
	}
	if rl.Allow("key2", 1, 100) {
		t.Fatal("key2 should be blocked")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(2, time.Minute)
	defer rl.Stop()

	router := gin.New()
	router.Use(RateLimitMiddleware(rl))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// First two should succeed
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// Third should be rate limited
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}
