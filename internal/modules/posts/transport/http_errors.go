package transport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	postApp "github.com/nextpresskit/backend/internal/modules/posts/application"
)

// respondPostsServiceError maps posts application errors to HTTP responses.
// It always writes a JSON body when err != nil. Returns whether err was non-nil.
func respondPostsServiceError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	switch {
	case errors.Is(err, postApp.ErrPostNotFound) || errors.Is(err, postApp.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
	case errors.Is(err, postApp.ErrInvalidPost),
		errors.Is(err, postApp.ErrInvalidSubresource),
		errors.Is(err, postApp.ErrInvalidArgument),
		errors.Is(err, postApp.ErrInvalidStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	case errors.Is(err, postApp.ErrSlugTaken) || errors.Is(err, postApp.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": msg})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
	return true
}
