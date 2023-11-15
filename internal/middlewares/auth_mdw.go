package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	userauth "github.com/khoaphungnguyen/go-openai/internal/user/auth"
)

// AuthMiddleware is a middleware that validates token and authorizes users
func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.GetHeader("Authorization")
		if clientToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		bearerToken := strings.TrimPrefix(clientToken, "Bearer ")
		if bearerToken == clientToken {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect Format of Authorization Token"})
			c.Abort()
			return
		}

		jwtWrapper := userauth.JwtWrapper{
			SecretKey: secretKey,
			Issuer:    "AuthService",
		}

		claims, err := jwtWrapper.ValidateToken(bearerToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Since Audience (Subject) in your JWT is the email, and not an ID,
		// it would be more appropriate to set the email in the context instead of converting it to an int.
		c.Set("email", claims.Subject)
		c.Next()
	}
}
