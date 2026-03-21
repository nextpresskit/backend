package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestFixedWindowRateLimiter_AllowsUpToLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewFixedWindowRateLimiter(2, time.Minute)

	r := gin.New()
	r.Use(limiter.Middleware("public"))
	r.GET("/v1/posts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/posts", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/posts", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", w.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed decoding response body: %v", err)
	}
	if body["error"] != "rate_limited" {
		t.Fatalf(`expected error "rate_limited", got %v`, body["error"])
	}
}

