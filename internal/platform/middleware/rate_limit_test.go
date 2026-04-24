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

func TestFixedWindowRateLimiter_ScopesAreIsolated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewFixedWindowRateLimiter(1, time.Minute)

	r := gin.New()
	r.Use(limiter.Middleware("public"))
	r.GET("/v1/public", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	admin := gin.New()
	admin.Use(limiter.Middleware("admin"))
	admin.GET("/v1/admin/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	req1 := httptest.NewRequest(http.MethodGet, "/v1/public", nil)
	req1.RemoteAddr = "1.2.3.4:1234"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first public request 200, got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil)
	req2.RemoteAddr = "1.2.3.4:1234"
	w2 := httptest.NewRecorder()
	admin.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected first admin request 200 due to scope isolation, got %d", w2.Code)
	}
}

func TestClientIPFromRequest_PrefersForwardedHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/posts", nil)
	req.RemoteAddr = "10.0.0.5:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.5")
	req.Header.Set("X-Real-IP", "198.51.100.2")

	got := clientIPFromRequest(req)
	if got != "203.0.113.10" {
		t.Fatalf("expected first forwarded IP, got %q", got)
	}
}

