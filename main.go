package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

var (
	client *openai.Client
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust the origin checking for production
	},
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// Get the OpenAI API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("The environment variable OPENAI_API_KEY is required")
		return
	}

	// Initialize OpenAI client
	client = openai.NewClient(apiKey)

	// Initialize the Gin router
	router := gin.Default()

	// CORS middleware for Gin
	router.Use(CORSMiddleware())

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})

	// Start the HTTP server
	fmt.Println("The server is started on port 8000")
	router.Run(":8000") // Defaults to ":8080" if not specified
}

// CORSMiddleware defines CORS policy for Gin
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Set to specific origin in production
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
		c.Writer.Header().Set("Content-Type", "application/json")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// wsHandler handles WebSocket connections
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error: %v", err)
			}
			break // Exit the loop; the connection is closed.
		}

		// Handle the message
		go chatHandler(client, string(message), conn)
	}
}

// chatHandler interacts with the OpenAI API and streams the response content
func chatHandler(client *openai.Client, userMessage string, conn *websocket.Conn) {
	// Set up the chat completion request
	req := openai.ChatCompletionRequest{
		Model:     "gpt-3.5-turbo-1106",
		MaxTokens: 100,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: userMessage,
			},
		},
	}

	// Create the chat completion stream
	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		log.Printf("CreateChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	// Stream the response from OpenAI to the WebSocket
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Println("Stream finished")
			} else {
				log.Printf("Stream error: %v\n", err)
			}
			break
		}

		// Send the response over the WebSocket
		if err := conn.WriteMessage(websocket.TextMessage, []byte(response.Choices[0].Delta.Content)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}
