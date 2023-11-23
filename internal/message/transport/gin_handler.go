package messagetransport

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	messagemodel "github.com/khoaphungnguyen/go-openai/internal/message/model"
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
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateThread handles the creation of a new chat thread.
func (mh *MessageHandler) CreateThread(c *gin.Context) {
	var payload ThreadPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	thread := &messagemodel.ChatThread{
		Title:  payload.Title,
		UserID: userID,
	}

	if err := mh.messsageService.CreateThread(thread); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to create thread")
		return
	}

	respondWithJSON(c, http.StatusCreated, convertToThreadResponse(thread))
}

// GetAllThreads handles the retrieval of all chat threads for a specific user.
func (mh *MessageHandler) GetAllThreads(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	threads, err := mh.messsageService.GetAllThreads(userID)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to retrieve threads")
		return
	}

	var responseThreads []ThreadResponse
	for _, thread := range threads {
		responseThreads = append(responseThreads, convertToThreadResponse(&thread))
	}

	respondWithJSON(c, http.StatusOK, responseThreads)
}

// GetThreadByID handles retrieving a single chat thread by its ID.
func (mh *MessageHandler) GetThreadByID(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid thread ID")
		return
	}

	thread, err := mh.messsageService.GetThreadByID(threadID)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error retrieving thread")
		return
	}

	// If thread is not found, return a not found error instead of forbidden
	if thread == nil {
		respondWithError(c, http.StatusNotFound, "Thread not found")
		return
	}

	// Check if the user is authorized to view the thread
	if thread.UserID != userID {
		respondWithError(c, http.StatusForbidden, "Access denied")
		return
	}

	respondWithJSON(c, http.StatusOK, convertToThreadResponse(thread))
}

// CreateMessage handles creating a new message in a chat thread.
func (mh *MessageHandler) CreateMessage(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	var message messagemodel.ChatMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid message format")
		return
	}

	message.UserID = userID
	if err := mh.messsageService.CreateMessage(userID, &message); err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(c, http.StatusCreated, convertToChatMessageResponse(&message))
}

// GetMessagesByThreadID handles retrieving messages for a specific thread with pagination.
func (mh *MessageHandler) GetMessagesByThreadID(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid thread ID")
		return
	}

	// Check if the user is authorized to access the thread
	if !mh.messsageService.IsUserThreadOwner(threadID, userID) {
		respondWithError(c, http.StatusForbidden, "Unauthorized access to thread")
		return
	}

	offset, limit := parseQueryInt(c, "offset", 0), parseQueryInt(c, "limit", 100)
	messages, err := mh.messsageService.GetMessagesByThreadID(threadID, offset, limit, userID)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to retrieve messages")
		return
	}
	respondWithJSON(c, http.StatusOK, messages)
}

// DeleteThread handles the deletion of a chat thread.
func (mh *MessageHandler) DeleteThread(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid thread ID")
		return
	}

	if err := mh.messsageService.DeleteThread(threadID, userID); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to delete thread")
		return
	}

	respondWithJSON(c, http.StatusOK, gin.H{"message": "Thread deleted successfully"})
}

// Helper functions
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, errors.New("userID not provided")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID")
	}

	return userID, nil
}

func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

func respondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

func convertToThreadResponse(thread *messagemodel.ChatThread) ThreadResponse {
	return ThreadResponse{
		ID:        thread.ID,
		Title:     thread.Title,
		CreatedAt: thread.CreatedAt,
		UpdatedAt: thread.UpdatedAt,
	}
}

// convertToChatMessageResponse converts a ChatMessage model to a ChatMessageResponse for the API.
func convertToChatMessageResponse(message *messagemodel.ChatMessage) ChatMessageResponse {
	if message == nil {
		return ChatMessageResponse{} // Return empty response for nil messages
	}

	return ChatMessageResponse{
		ID:        message.ID,
		ThreadID:  message.ThreadID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}
}

// parseQueryInt tries to parse an integer from query parameters.
func parseQueryInt(c *gin.Context, param string, defaultValue int) int {
	valueStr := c.DefaultQuery(param, fmt.Sprintf("%d", defaultValue))
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
