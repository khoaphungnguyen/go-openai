package chatbusiness

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	chatmodel "github.com/khoaphungnguyen/go-openai/internal/chat/model"
)

// CreateThread handles the creation of a new chat thread.
func (cs *ChatService) CreateThread(thread *chatmodel.ChatThread) error {
	if thread == nil {
		return errors.New("thread cannot be nil")
	}
	return cs.chatStore.CreateThread(thread)
}

// GetThreadByID retrieves a chat thread by its ID.
func (cs *ChatService) GetThreadByID(threadID uuid.UUID) (*chatmodel.ChatThread, error) {
	if threadID == uuid.Nil {
		return nil, errors.New("invalid thread ID")
	}
	return cs.chatStore.GetThreadByID(threadID)
}

// GetThreadsByUserID retrieves all chat threads for a specific user.
func (cs *ChatService) GetThreadsByUserID(userID uuid.UUID, limit, offset int) ([]chatmodel.ChatThread, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}
	return cs.chatStore.GetThreadsByUserID(userID, limit, offset)
}

// CreateMessage adds a new message to a chat thread.
func (cs *ChatService) CreateMessage(userID uuid.UUID, message *chatmodel.ChatMessage) error {
	if message == nil {
		return errors.New("message cannot be nil")
	}
	exists, err := cs.chatStore.CheckThreadExistsAndBelongsToUser(message.ThreadID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify thread ownership: %w", err)
	}
	if !exists {
		return errors.New("thread does not exist or user lacks permission")
	}
	return cs.chatStore.CreateMessage(message)
}

// GetAllThreads retrieves all chat threads for a specific user.
func (cs *ChatService) GetAllThreads(userID uuid.UUID) ([]chatmodel.ChatThread, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}
	return cs.chatStore.GetAllThreads(userID)
}

// GetMessagesByThreadID retrieves messages of a chat thread with pagination.
func (cs *ChatService) GetMessagesByThreadID(threadID uuid.UUID, limit, offset int, userID uuid.UUID) ([]chatmodel.ChatMessage, error) {
	if threadID == uuid.Nil {
		return nil, errors.New("invalid thread ID")
	}
	if !cs.chatStore.IsUserThreadOwner(threadID, userID) {
		return nil, errors.New("unauthorized access to thread")
	}
	return cs.chatStore.GetMessagesByThreadID(threadID, limit, offset)
}

// DeleteThread deletes a chat thread.
func (cs *ChatService) DeleteThread(threadID uuid.UUID, userID uuid.UUID) error {
	if threadID == uuid.Nil {
		return errors.New("invalid thread ID")
	}
	if !cs.chatStore.IsUserThreadOwner(threadID, userID) {
		return errors.New("unauthorized access to thread")
	}
	return cs.chatStore.DeleteThread(threadID, userID)
}

// CheckThreadExists verifies if a thread exists in the database.
func (cs *ChatService) CheckThreadExists(threadID uuid.UUID) (bool, error) {
	if threadID == uuid.Nil {
		return false, errors.New("invalid thread ID")
	}
	return cs.chatStore.CheckThreadExists(threadID)
}

func (cs *ChatService) IsUserThreadOwner(threadID, userID uuid.UUID) bool {
	return cs.chatStore.IsUserThreadOwner(threadID, userID)
}
