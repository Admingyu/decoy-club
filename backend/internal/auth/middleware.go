package auth

import (
	"net/http"
	"strings"

	"decoy-club/backend/internal/common/response"

	"github.com/gin-gonic/gin"
)

const ContextKeyClaims = "auth.claims"

func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			c.Abort()
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))
		token = strings.TrimSpace(token)
		if token == "" {
			response.JSON(c, http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			c.Abort()
			return
		}

		claims, err := VerifyJWT(secret, token)
		if err != nil {
			response.JSON(c, http.StatusUnauthorized, gin.H{"error": "invalid authorization token"})
			c.Abort()
			return
		}

		c.Set(ContextKeyClaims, claims)
		c.Next()
	}
}
