package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

func OpenAIClientMiddleware(client *openai.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil {
			log.Println("OpenAI client has not been initialized")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		// Set the OpenAI client in the context
		c.Set("openaiClient", client)

		// Proceed to the next middleware/handler
		c.Next()
	}
}
