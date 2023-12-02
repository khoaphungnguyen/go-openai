package openaitransport

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

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

// Define a map to hold channels for each thread
var threadSSEChannels = make(map[uuid.UUID]chan string)

var mutex = &sync.Mutex{}

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
		Stream:    true,
		MaxTokens: 1000,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: inputData.Message,
			},
		},
	}
	stream, err := openaiClient.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		log.Printf("CreateChatCompletionStream error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stream"})
		return
	}
	defer stream.Close()

	var responseBuilder strings.Builder
	// Stream the response from OpenAI and send parts to the client via SSE
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
		// Send each part of the response to the SSE channel
		mutex.Lock()
		if ch, exists := threadSSEChannels[threadID]; exists {
			ch <- responseContent
		}
		mutex.Unlock()
	}

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

	c.JSON(http.StatusOK, gin.H{"message": "Message received and processed"})
}

func (h *OpenAIHandler) SSEHandler(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, "Invalid thread ID")
		return
	}
	// Initialize a channel for the thread if it doesn't exist
	mutex.Lock()
	if _, exists := threadSSEChannels[threadID]; !exists {
		threadSSEChannels[threadID] = make(chan string, 50) // Adjust buffer size as needed
	}
	ch := threadSSEChannels[threadID]
	mutex.Unlock()

	// Set SSE Headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		common.RespondWithError(c, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	// Stream messages for the specific thread
	for {
		select {
		case response := <-ch:
			jsonResponse := fmt.Sprintf(`{"Content": %q, "Role": "assistant"}`, response)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonResponse)
			flusher.Flush()

		case <-c.Request.Context().Done():
			// Cleanup when client disconnects
			mutex.Lock()
			close(ch)
			delete(threadSSEChannels, threadID)
			mutex.Unlock()
			return
		}
	}
}

// FetchSuggestion handles the request to fetch suggestions from OpenAI.
func (h *OpenAIHandler) FetchSuggestion(c *gin.Context) {
	// Extract user ID from context, if required
	_, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Retrieve the OpenAI client from the context
	openaiClient, exists := c.MustGet("openaiClient").(*openai.Client)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI client not available"})
		return
	}
	type RequestData struct {
		Model string `json:"model"`
	}

	// Bind the input data (assumed to be the model name)
	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Construct the prompt
	prompt := `Provide four engaging recommendations (max 10 words each) as JSON: [{ "title": "", "content": "" }, ...]`

	// Create the request payload and send the request to OpenAI
	resp, err := openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:            requestData.Model,
			Messages:         []openai.ChatCompletionMessage{{Role: "user", Content: prompt}},
			Temperature:      0.7,
			TopP:             1,
			FrequencyPenalty: 0,
			PresencePenalty:  0,
			MaxTokens:        1000,
			N:                1,
			Stream:           false,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
		return
	}

	// Check if the response has content and return the content
	if len(resp.Choices) > 0 && len(resp.Choices[0].Message.Content) > 0 {
		c.JSON(http.StatusOK, resp.Choices[0].Message.Content)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "No content in response"})
}
