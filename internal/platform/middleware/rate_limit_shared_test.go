package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type counterStoreStub struct {
	count int64
}

func (s *counterStoreStub) IncrementWindow(_ context.Context, _ string, _ time.Duration) (int64, time.Duration, error) {
	s.count++
	return s.count, time.Minute, nil
}

func TestSharedFixedWindowRateLimiter_BlocksAfterLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &counterStoreStub{}
	limiter := NewSharedFixedWindowRateLimiter(1, time.Minute, store)

	r := gin.New()
	r.Use(limiter.Middleware("public"))
	r.GET("/posts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req1 := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req2.RemoteAddr = "1.2.3.4:1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d", w2.Code)
	}
}
