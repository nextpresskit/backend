package middleware

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// FixedWindowRateLimiter implements an in-memory fixed-window rate limiter.
// It is safe for multi-request concurrency in a single process.
//
// Note: For multi-instance deployments you will eventually want a shared
// store (e.g. Redis). Phase 4 uses an in-memory implementation for simplicity.
type FixedWindowRateLimiter struct {
	maxRequests int
	window      time.Duration

	mu      sync.Mutex
	buckets map[string]*rateBucket
}

type rateBucket struct {
	windowStart time.Time
	count       int
}

type SharedFixedWindowRateLimiter struct {
	maxRequests int
	window      time.Duration
	store       RateLimitCounterStore
}

func NewFixedWindowRateLimiter(maxRequests int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		maxRequests: maxRequests,
		window:      window,
		buckets:     make(map[string]*rateBucket),
	}
}

func NewSharedFixedWindowRateLimiter(maxRequests int, window time.Duration, store RateLimitCounterStore) *SharedFixedWindowRateLimiter {
	return &SharedFixedWindowRateLimiter{
		maxRequests: maxRequests,
		window:      window,
		store:       store,
	}
}

// Middleware creates a Gin middleware that rate limits per client IP and scope.
// If maxRequests <= 0, the middleware becomes a no-op.
func (r *FixedWindowRateLimiter) Middleware(scope string) gin.HandlerFunc {
	if r.maxRequests <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		clientIP := clientIPFromRequest(c.Request)
		if clientIP == "" {
			clientIP = "unknown"
		}

		key := fmt.Sprintf("%s|%s", clientIP, scope)
		now := time.Now().UTC()

		var retryAfterSeconds int
		var allowed bool

		r.mu.Lock()
		b := r.buckets[key]
		if b == nil {
			b = &rateBucket{windowStart: now, count: 0}
			r.buckets[key] = b
		}

		// Reset bucket if window elapsed.
		if now.Sub(b.windowStart) >= r.window {
			b.windowStart = now
			b.count = 0
		}

		if b.count >= r.maxRequests {
			remaining := r.window - now.Sub(b.windowStart)
			if remaining < 0 {
				remaining = 0
			}
			// Avoid zero due to rounding; client experience is better.
			retryAfterSeconds = int(math.Ceil(remaining.Seconds()))
			allowed = false
		} else {
			b.count++
			allowed = true
		}
		r.mu.Unlock()

		if !allowed {
			if retryAfterSeconds > 0 {
				c.Writer.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
			}
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":                  "rate_limited",
				"retryAfterSeconds":      retryAfterSeconds,
				"rateLimitScope":         scope,
				"rateLimitWindowSeconds": int(r.window.Seconds()),
				"rateLimitMax":           r.maxRequests,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Middleware creates a Gin middleware that rate limits per client IP and scope
// with a shared external counter store (e.g. Redis).
func (r *SharedFixedWindowRateLimiter) Middleware(scope string) gin.HandlerFunc {
	if r.maxRequests <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		clientIP := clientIPFromRequest(c.Request)
		if clientIP == "" {
			clientIP = "unknown"
		}
		key := fmt.Sprintf("%s|%s", clientIP, scope)
		count, ttl, err := r.store.IncrementWindow(c.Request.Context(), key, r.window)
		if err != nil {
			// Fail-open on store errors to keep request path available.
			c.Next()
			return
		}
		if count > int64(r.maxRequests) {
			retryAfterSeconds := int(math.Ceil(ttl.Seconds()))
			if retryAfterSeconds > 0 {
				c.Writer.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
			}
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":                  "rate_limited",
				"retryAfterSeconds":      retryAfterSeconds,
				"rateLimitScope":         scope,
				"rateLimitWindowSeconds": int(r.window.Seconds()),
				"rateLimitMax":           r.maxRequests,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func clientIPFromRequest(r *http.Request) string {
	// Prefer X-Forwarded-For as it is the most common in real deployments.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return firstHost(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return firstHost(xri)
	}
	// Fall back to remote addr.
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func firstHost(x string) string {
	// XFF can contain multiple IPs separated by commas.
	for i := 0; i < len(x); i++ {
		if x[i] == ',' {
			return strings.TrimSpace(x[:i])
		}
	}
	return strings.TrimSpace(x)
}
