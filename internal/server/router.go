package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gorm.io/gorm"

	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
)

type ReadinessCheck struct {
	Name  string
	Check func(context.Context) error
}

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nextpresskit_http_requests_total",
			Help: "Total HTTP requests handled by the API.",
		},
		[]string{"method", "route", "status"},
	)
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nextpresskit_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)
)

// ConfigureEngine applies global middleware and registers all top-level routes.
// Feature modules will later plug into the provided router via dedicated
// registration functions to keep boundaries clear.
func ConfigureEngine(engine *gin.Engine, log *zap.SugaredLogger, db *gorm.DB, appVersion string, checks ...ReadinessCheck) {
	// In production you typically want to disable Gin's debug output and rely
	// on structured logging instead.
	gin.SetMode(gin.ReleaseMode)

	// Global middleware stack. We keep this minimal in Phase 1 and will extend
	// it (e.g. for authentication, request IDs, metrics) in later phases.
	engine.Use(cors.New(buildCORSConfig()))
	engine.Use(gin.Recovery())
	engine.Use(platformMiddleware.RequestIDMiddleware())
	engine.Use(requestLoggingMiddleware(log))
	engine.Use(metricsMiddleware())
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": appVersion,
		})
	})

	engine.GET("/ready", func(c *gin.Context) {
		// A lightweight database check ensures we only report readiness when
		// core dependencies are available. This keeps the handler cheap while
		// still being meaningful for load balancers and orchestrators.
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not ready",
				"details": "database handle not initialized",
			})
			return
		}

		if err := db.Exec("SELECT 1").Error; err != nil {
			log.Warnw("readiness check failed",
				"component", "database",
				"error", err,
			)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not ready",
				"details": "database check failed",
			})
			return
		}

		for _, check := range checks {
			if check.Check == nil {
				continue
			}
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			err := check.Check(ctx)
			cancel()
			if err != nil {
				log.Warnw("readiness check failed",
					"component", check.Name,
					"error", err,
				)
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status":    "not ready",
					"component": check.Name,
					"details":   "dependency check failed",
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ready",
			"version": appVersion,
		})
	})

	// Root endpoint to confirm that the service is running and identify the
	// backend explicitly (useful in multi-service environments).
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "nextpresskit-backend",
			"version": appVersion,
		})
	})
}

// requestLoggingMiddleware provides concise, structured logging of incoming
// HTTP traffic. It avoids duplicating Gin's own debug logging while still
// emitting meaningful data for production troubleshooting.
func requestLoggingMiddleware(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		var requestID string
		if v, ok := c.Get(platformMiddleware.ContextRequestIDKey); ok {
			requestID, _ = v.(string)
		}

		var userID string
		if v, ok := c.Get(platformMiddleware.ContextUserIDKey); ok {
			userID, _ = v.(string)
		}

		c.Next()

		status := c.Writer.Status()
		latencyMs := time.Since(start).Milliseconds()
		log.Infow("http request completed",
			"method", method,
			"path", path,
			"status", status,
			"client_ip", clientIP,
			"request_id", requestID,
			"user_id", userID,
			"latency_ms", latencyMs,
		)
	}
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		httpRequestsTotal.WithLabelValues(method, route, status).Inc()
		httpRequestDuration.WithLabelValues(method, route, status).Observe(time.Since(start).Seconds())
	}
}

