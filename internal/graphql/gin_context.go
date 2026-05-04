package graphql

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ginContextKey struct{}

// WithGinContext attaches the current *gin.Context to the request context so
// GraphQL resolvers can set cookies/headers.
func WithGinContext(ctx context.Context, c *gin.Context) context.Context {
	return context.WithValue(ctx, ginContextKey{}, c)
}

func GinContextFrom(ctx context.Context) (*gin.Context, bool) {
	c, ok := ctx.Value(ginContextKey{}).(*gin.Context)
	return c, ok && c != nil
}

