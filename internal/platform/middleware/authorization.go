package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	rbacDomain "github.com/nextpresskit/backend/internal/modules/rbac/domain"
)

func RequirePermission(checker rbacDomain.PermissionChecker, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(ContextUserIDKey)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_user_context"})
			c.Abort()
			return
		}

		userID, _ := v.(string)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_user_context"})
			c.Abort()
			return
		}

		allowed, err := checker.UserHasPermission(c.Request.Context(), userID, permissionCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_check_failed"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}

