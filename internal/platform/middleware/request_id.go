package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContextRequestIDKey stores the request correlation id in the Gin context.
// It can be used by downstream handlers and for structured logging.
const ContextRequestIDKey = "request_id"

// RequestIDMiddleware ensures each request has a correlation id.
// - If `X-Request-ID` header is provided, it is reused.
// - Otherwise, a UUID is generated.
// - The id is stored in Gin context and echoed back via response header.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
		}

		c.Set(ContextRequestIDKey, rid)
		c.Writer.Header().Set("X-Request-ID", rid)

		// Ensure we don't accidentally write an empty header value.
		if c.Writer.Header().Get(http.CanonicalHeaderKey("X-Request-ID")) == "" {
			c.Writer.Header().Set("X-Request-ID", rid)
		}

		c.Next()
	}
}

