package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	userauth "github.com/khoaphungnguyen/go-openai/internal/user/auth"
)

const bearerSchema = "Bearer "

// AuthMiddleware is a middleware that validates JWT tokens and authorizes users.
func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			return
		}

		if !strings.HasPrefix(authHeader, bearerSchema) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerSchema)
		jwtWrapper := userauth.JwtWrapper{SecretKey: secretKey}

		claims, err := jwtWrapper.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Add userID to Gin context for downstream handlers
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
