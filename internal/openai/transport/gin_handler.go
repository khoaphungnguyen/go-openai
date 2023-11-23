package openaitransport

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khoaphungnguyen/go-openai/internal/common"
	openaimodel "github.com/khoaphungnguyen/go-openai/internal/openai/model"
	"github.com/sashabaranov/go-openai"
)

// CreateTransaction handles the creation of a new OpenAI transaction (HTTP Handler).
func (h *OpenAIHandler) CreateTransaction(c *gin.Context) {
	userID, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	var inputData openaimodel.OpenAITransactionInput
	if err := c.ShouldBindJSON(&inputData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err = h.createTransaction(userID, inputData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction created"})
}

// Core logic for creating a transaction
func (h *OpenAIHandler) createTransaction(userID uuid.UUID, inputData openaimodel.OpenAITransactionInput) error {
	threadID, err := uuid.Parse(inputData.ThreadID)
	if err != nil {
		return err
	}

	return h.openAIService.CreateTransaction(userID, threadID, inputData.Message, inputData.Model, inputData.Role)
}

// GetTransactionsByUserID handles fetching transactions for a specific user.
func (h *OpenAIHandler) GetTransactionsByUserID(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	transactions, err := h.openAIService.GetTransactionsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// UpdateTransaction handles the updating of an existing OpenAI transaction.
func (h *OpenAIHandler) UpdateTransaction(c *gin.Context) {
	var transaction openaimodel.OpenAITransaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.openAIService.UpdateTransaction(&transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction updated"})
}

// DeleteTransaction handles the deletion of an OpenAI transaction.
func (h *OpenAIHandler) DeleteTransaction(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("transactionID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	if err := h.openAIService.DeleteTransaction(transactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted"})
}

// GetTransactionByID handles fetching a specific transaction by its ID.
func (h *OpenAIHandler) GetTransactionByID(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("transactionID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	transaction, err := h.openAIService.GetTransactionByID(transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func (h *OpenAIHandler) MessageHanlder(c *gin.Context) {
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

	openaiClient, exists := c.MustGet("openaiClient").(*openai.Client)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI client not available"})
		return
	}

	var inputData openaimodel.OpenAITransactionInput
	if err := c.ShouldBindJSON(&inputData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	// Log the user's message
	err = h.createTransaction(userID, openaimodel.OpenAITransactionInput{
		ThreadID: threadID.String(),
		Message:  string(inputData.Message),
		Model:    inputData.Model,
		Role:     "user",
	})
	if err != nil {
		log.Printf("Error saving user transaction: %v", err)
		return
	}
	// Set up the chat completion request using the OpenAI client
	req := openai.ChatCompletionRequest{
		Model:     inputData.Model,
		MaxTokens: 100,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: inputData.Message,
			},
		},
	}
	fmt.Println("Got here. AI...1")
	stream, err := openaiClient.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		log.Printf("CreateChatCompletionStream error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stream"})
		return
	}

	var responseBuilder strings.Builder
	// Stream the response from OpenAI
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Stream error: %v\n", err)
			break
		}
		responseContent := response.Choices[0].Delta.Content
		responseBuilder.WriteString(responseContent)
	}
	stream.Close()

	aiResponse := responseBuilder.String()

	// Log AI response as a transaction
	err = h.createTransaction(userID, openaimodel.OpenAITransactionInput{
		ThreadID: threadID.String(),
		Message:  responseBuilder.String(),
		Model:    inputData.Model,
		Role:     "assistant",
	})
	if err != nil {
		log.Printf("Error saving user transaction: %v", err)
		return
	}
	aiResponseQueue <- aiResponse // Send AI response to the SSE queue
	c.JSON(http.StatusOK, gin.H{"message": "Message received and processed"})
}

var aiResponseQueue = make(chan string, 200)

func (h *OpenAIHandler) SSEHandler(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	fmt.Println("SSE begin...")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		common.RespondWithError(c, http.StatusInternalServerError, "Streaming not supported")
		return
	}
	fmt.Println("SSE", aiResponseQueue)
	for {
		select {
		case response := <-aiResponseQueue:
			// Create a JSON object that includes both the content and the role.
			// You may need to adjust this depending on how your data is structured.
			fmt.Println("response", response)
			jsonResponse := fmt.Sprintf(`{"Content": %q, "Role": "assistant"}`, response)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonResponse)
			flusher.Flush()
		case <-c.Request.Context().Done():
			return

		}
	}
}
