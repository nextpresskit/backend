package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nextpresskit/backend/internal/config"
)

type AccessTokenParser interface {
	ParseAccessToken(token string) (string, error)
}

const ContextUserIDKey = "user_id"

func AuthRequired(parser AccessTokenParser, jwtCfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		if strings.EqualFold(strings.TrimSpace(jwtCfg.AuthSource), "header") {
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
			token = parts[1]
		} else {
			ck, err := c.Request.Cookie(jwtCfg.AccessCookieName)
			if err != nil || ck == nil || strings.TrimSpace(ck.Value) == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_access_token_cookie"})
				c.Abort()
				return
			}
			token = ck.Value
		}

		userID, err := parser.ParseAccessToken(token)
		if err != nil || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_or_expired_token"})
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}
