package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware defines CORS policy for Gin
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if the origin is in the list of allowed origins
		allowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // or specify for production
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type") // Add any other headers you want to expose

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			log.Println("Preflight request from origin:", origin)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
