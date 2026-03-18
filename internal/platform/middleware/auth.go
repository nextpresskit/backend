package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AccessTokenParser interface {
	ParseAccessToken(token string) (string, error)
}

const ContextUserIDKey = "user_id"

func AuthRequired(parser AccessTokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_authorization_header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_authorization_header"})
			c.Abort()
			return
		}

		userID, err := parser.ParseAccessToken(parts[1])
		if err != nil || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_or_expired_token"})
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}
