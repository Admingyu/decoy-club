package middleware

import "github.com/gin-gonic/gin"

func CORS(frontendOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if frontendOrigin != "" {
			c.Header("Access-Control-Allow-Origin", frontendOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Credentials", "false")
		}
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
