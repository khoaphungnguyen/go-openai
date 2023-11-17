package chatgin

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	chatmodel "github.com/khoaphungnguyen/go-openai/internal/chat/model"
)

type ThreadPayload struct {
	Title string `json:"title"`
}

type ThreadResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ChatMessageResponse struct {
	ID        uuid.UUID `json:"id"`
	ThreadID  uuid.UUID `json:"threadId"`
	Role      string    `json:"role"`
	Model     string    `json:"model"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateThread handles creating a new chat thread
func (ch *ChatHandler) CreateThread(c *gin.Context) {
	var payload ThreadPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not provided"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	thread := &chatmodel.ChatThread{
		Title:  payload.Title,
		UserID: userID,
	}

	if err := ch.chatService.CreateThread(thread); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create a response object
	response := ThreadResponse{
		ID:        thread.ID,
		Title:     thread.Title,
		CreatedAt: thread.CreatedAt,
		UpdatedAt: thread.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetAllThreads retrieves a list of all chat threads
func (ch *ChatHandler) GetAllThreads(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not provided"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	threads, err := ch.chatService.GetAllThreads(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert each thread to the response format
	var responseThreads []ThreadResponse
	for _, thread := range threads {
		responseThread := ThreadResponse{
			ID:        thread.ID,
			Title:     thread.Title,
			CreatedAt: thread.CreatedAt,
			UpdatedAt: thread.UpdatedAt,
		}
		responseThreads = append(responseThreads, responseThread)
	}

	c.JSON(http.StatusOK, responseThreads)
}

// GetThread handles retrieving a chat thread by its ID
func (ch *ChatHandler) GetThread(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not provided"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	threadIDStr := c.Param("id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	// Check if the thread exists
	exists, err = ch.chatService.CheckThreadExists(threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking thread existence"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"message": "Thread not found or already deleted"})
		return
	}

	// Retrieve the thread
	thread, err := ch.chatService.GetThreadByID(threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving thread"})
		return
	}

	if thread.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	responseThread := ThreadResponse{
		ID:        thread.ID,
		Title:     thread.Title,
		CreatedAt: thread.CreatedAt,
		UpdatedAt: thread.UpdatedAt,
	}

	c.JSON(http.StatusOK, responseThread)
}

// CreateMessage handles creating a new message in a chat thread
func (ch *ChatHandler) CreateMessage(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not provided"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	var message chatmodel.ChatMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	message.UserID = userID
	if err := ch.chatService.CreateMessage(userID, &message); err != nil {
        if err.Error() == "thread does not exist or you do not have permission to post in it" {
            c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

	// Create a response object
	response := ChatMessageResponse{
		ID:        message.ID,
		ThreadID:  message.ThreadID,
		Role:      message.Role,
		Model:     message.Model,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetMessagesByThreadID handles the retrieval of chat messages for a specific thread with pagination
func (ch *ChatHandler) GetMessagesByThreadID(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not provided"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	threadIDStr := c.Param("threadID")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	// Parse pagination parameters from the query, with defaults
	offset := parseQueryInt(c, "offset", 0)
	limit := parseQueryInt(c, "limit", 10)

	messages, err := ch.chatService.GetMessagesByThreadID(threadID, offset, limit, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// Helper function to parse integer query parameters
func parseQueryInt(c *gin.Context, param string, defaultValue int) int {
	valueStr := c.DefaultQuery(param, fmt.Sprintf("%d", defaultValue))
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// DeleteThread handles the hard deletion of a chat thread
func (ch *ChatHandler) DeleteThread(c *gin.Context) {

	threadIDStr := c.Param("id")
	threadID, err := uuid.Parse(threadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid thread ID"})
		return
	}

	// Check if the thread exists before attempting to delete
	exists, err := ch.chatService.CheckThreadExists(threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if thread exists"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Thread not found or already deleted"})
		return
	}
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not provided"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Proceed with deletion
	if err := ch.chatService.DeleteThread(threadID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete thread"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Thread deleted successfully"})
}
