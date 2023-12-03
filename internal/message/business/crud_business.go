package messagebusiness

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	messagemodel "github.com/khoaphungnguyen/go-openai/internal/message/model"
)

// CreateThread handles the creation of a new chat thread.
func (ms *MessageService) CreateThread(thread *messagemodel.ChatThread) error {
	if thread == nil {
		return errors.New("thread cannot be nil")
	}
	return ms.messageStore.CreateThread(thread)
}

// GetThreadByID retrieves a chat thread by its ID.
func (ms *MessageService) GetThreadByID(threadID uuid.UUID) (*messagemodel.ChatThread, error) {
	if threadID == uuid.Nil {
		return nil, errors.New("invalid thread ID")
	}
	return ms.messageStore.GetThreadByID(threadID)
}

// GetThreadsByUserID retrieves all chat threads for a specific user.
func (ms *MessageService) GetThreadsByUserID(userID uuid.UUID, limit, offset int) ([]messagemodel.ChatThread, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}
	return ms.messageStore.GetThreadsByUserID(userID, limit, offset)
}

// CreateMessage adds a new message to a chat thread.
func (ms *MessageService) CreateMessage(userID uuid.UUID, message *messagemodel.ChatMessage) error {
	if message == nil {
		return errors.New("message cannot be nil")
	}
	exists, err := ms.messageStore.CheckThreadExistsAndBelongsToUser(message.ThreadID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify thread ownership: %w", err)
	}
	if !exists {
		return errors.New("thread does not exist or user lacks permission")
	}
	return ms.messageStore.CreateMessage(message)
}

// GetAllThreads retrieves all chat threads for a specific user.
func (ms *MessageService) GetAllThreads(userID uuid.UUID) ([]messagemodel.ChatThread, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}
	return ms.messageStore.GetAllThreads(userID)
}

// GetMessagesByThreadID retrieves messages of a chat thread with pagination.
func (ms *MessageService) GetMessagesByThreadID(threadID uuid.UUID, limit, offset int, userID uuid.UUID) ([]messagemodel.ChatMessageResponse, error) {
	if threadID == uuid.Nil {
		return nil, errors.New("invalid thread ID")
	}
	messages, err := ms.messageStore.GetMessagesByThreadID(threadID, limit, offset)
	if err != nil {
		return nil, err
	}
	var responseMessages []messagemodel.ChatMessageResponse
	for _, msg := range messages {
		responseMessage := messagemodel.ChatMessageResponse{
			ID : msg.ID,
			Content:   msg.Content,
			Role:      msg.Role,
			CreatedAt: msg.CreatedAt,
		}
		responseMessages = append(responseMessages, responseMessage)
	}
	return responseMessages, nil
}

// DeleteThread deletes a chat thread.
func (ms *MessageService) DeleteThread(threadID uuid.UUID, userID uuid.UUID) error {
	if threadID == uuid.Nil {
		return errors.New("invalid thread ID")
	}
	if !ms.messageStore.IsUserThreadOwner(threadID, userID) {
		return errors.New("unauthorized access to thread")
	}
	return ms.messageStore.DeleteThread(threadID, userID)
}

// CheckThreadExists verifies if a thread exists in the database.
func (ms *MessageService) CheckThreadExists(threadID uuid.UUID) (bool, error) {
	if threadID == uuid.Nil {
		return false, errors.New("invalid thread ID")
	}
	return ms.messageStore.CheckThreadExists(threadID)
}

func (ms *MessageService) IsUserThreadOwner(threadID, userID uuid.UUID) bool {
	return ms.messageStore.IsUserThreadOwner(threadID, userID)
}
