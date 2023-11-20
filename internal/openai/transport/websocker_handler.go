package openaitransport

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/khoaphungnguyen/go-openai/internal/common"
	openaimodel "github.com/khoaphungnguyen/go-openai/internal/openai/model"
	"github.com/sashabaranov/go-openai"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Update for production
	},
}

// WebSocketHandler is the Gin handler for WebSocket connections
func (h *OpenAIHandler) WebSocketHandler(c *gin.Context) {
	// Extract the OpenAI client
	openaiClient, exists := c.MustGet("openaiClient").(*openai.Client)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI client not available"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	userID, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, "Invalid thread ID")
		return
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Log the user's message
		err = h.createTransaction(userID, openaimodel.OpenAITransactionInput{
			ThreadID: threadID.String(),
			Message:  string(message),
			Model:    "gpt-3.5-turbo",
			Role:     "user",
		})

		if err != nil {
			log.Printf("Error saving user transaction: %v", err)
			continue
		}

		// Set up the chat completion request using the OpenAI client
		req := openai.ChatCompletionRequest{
			Model:     "gpt-3.5-turbo-1106",
			MaxTokens: 100,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "user",
					Content: string(message),
				},
			},
		}

		stream, err := openaiClient.CreateChatCompletionStream(context.Background(), req)
		if err != nil {
			log.Printf("CreateChatCompletionStream error: %v\n", err)
			continue
		}
		saveSteam := ""
		// Stream the response from OpenAI
		for {
			response, err := stream.Recv()
			if err == io.EOF {
				//log.Println("Stream finished")
				break
			} else if err != nil {
				log.Printf("Stream error: %v\n", err)
				break
			}

			responseContent := response.Choices[0].Delta.Content
			saveSteam += responseContent
			// Send the response over the WebSocket
			if err := conn.WriteMessage(websocket.TextMessage, []byte(responseContent)); err != nil {
				log.Println("Write error:", err)
				return
			}
		}
		stream.Close()
		// Log the AI's response
		err = h.createTransaction(userID, openaimodel.OpenAITransactionInput{
			ThreadID: threadID.String(),
			Message:  saveSteam,
			Model:    "gpt-3.5-turbo",
			Role:     "assistant",
		})
		if err != nil {
			log.Printf("Error saving AI's response: %v", err)
			continue
		}

	}
}
