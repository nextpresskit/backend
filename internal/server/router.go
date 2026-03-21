package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	platformMiddleware "github.com/Petar-V-Nikolov/nextpress-backend/internal/platform/middleware"
)

// ConfigureEngine applies global middleware and registers all top-level routes.
// Feature modules will later plug into the provided router via dedicated
// registration functions to keep boundaries clear.
func ConfigureEngine(engine *gin.Engine, log *zap.SugaredLogger, db *gorm.DB) {
	// In production you typically want to disable Gin's debug output and rely
	// on structured logging instead.
	gin.SetMode(gin.ReleaseMode)

	// Global middleware stack. We keep this minimal in Phase 1 and will extend
	// it (e.g. for authentication, request IDs, metrics) in later phases.
	engine.Use(gin.Recovery())
	engine.Use(platformMiddleware.RequestIDMiddleware())
	engine.Use(requestLoggingMiddleware(log))

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
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

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	// Root endpoint to confirm that the service is running and identify the
	// backend explicitly (useful in multi-service environments).
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "nextpress-backend",
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

