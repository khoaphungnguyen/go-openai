package openaitransport

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sashabaranov/go-openai"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Update for production
	},
}

// WebSocketHandler is the Gin handler for WebSocket connections
func (h *OpenAIHandler) WebSocketHandler(c *gin.Context) {
	// Extract the OpenAI client and other necessary data from Gin context
	openaiClient, exists := c.MustGet("openaiClient").(*openai.Client)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI client not available"})
		return
	}

	w := c.Writer
	r := c.Request

	// Rest of the WebSocket upgrade process
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Pass the OpenAI client to the handler that processes each message
		go handleChatInteraction(openaiClient, string(message), conn)
	}
}

func handleChatInteraction(client *openai.Client, userMessage string, conn *websocket.Conn) {
	// Set up the chat completion request using the OpenAI client
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
