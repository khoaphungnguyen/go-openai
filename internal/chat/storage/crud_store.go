package chatstorage

import (
	"fmt"

	"github.com/google/uuid"
	chatmodel "github.com/khoaphungnguyen/go-openai/internal/chat/model"
)

// CreateThread adds a new chat thread to the database
func (cs *chatStore) CreateThread(thread *chatmodel.ChatThread) error {
	return cs.db.Create(thread).Error
}

// GetThreadByID retrieves a chat thread by its ID
func (cs *chatStore) GetThreadByID(threadID uuid.UUID) (*chatmodel.ChatThread, error) {
	var thread chatmodel.ChatThread

	// Check if the thread exists
	exists, err := cs.CheckThreadExists(threadID)
	if err != nil {
		return nil, fmt.Errorf("error checking if thread exists: %w", err)
	}
	if !exists {
		return nil, nil // Return nil if the thread does not exist
	}

	err = cs.db.Where("id = ?", threadID).First(&thread).Error
	if err != nil {
		return nil, err // Handle other errors (e.g., database errors)
	}

	return &thread, nil
}

// GetThreadsByUserID retrieves all chat threads for a specific user
func (cs *chatStore) GetThreadsByUserID(userID uuid.UUID, limit, offset int) ([]chatmodel.ChatThread, error) {
	var threads []chatmodel.ChatThread
	err := cs.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&threads).Error
	return threads, err
}

// GetAllThreads retrieves all chat threads for a specific user
func (cs *chatStore) GetAllThreads(userID uuid.UUID) ([]chatmodel.ChatThread, error) {
	var threads []chatmodel.ChatThread
	err := cs.db.Where("user_id = ?", userID).Find(&threads).Error
	return threads, err
}

// CreateMessage adds a new message to a chat thread
func (cs *chatStore) CreateMessage(message *chatmodel.ChatMessage) error {
	return cs.db.Create(message).Error
}

func (cs *chatStore) IsUserThreadOwner(threadID, userID uuid.UUID) bool {
	var count int64
	cs.db.Model(&chatmodel.ChatThread{}).Where("id = ? AND user_id = ?", threadID, userID).Count(&count)
	return count > 0
}

// GetMessagesByThreadID retrieves messages of a chat thread with pagination
func (cs *chatStore) GetMessagesByThreadID(threadID uuid.UUID, offset, limit int) ([]chatmodel.ChatMessage, error) {
	var messages []chatmodel.ChatMessage
	result := cs.db.Where("thread_id = ?", threadID).
		Offset(offset).
		Limit(limit).
		Find(&messages)
	return messages, result.Error
}

func (cs *chatStore) DeleteThread(threadID uuid.UUID, userID uuid.UUID) error {
	return cs.db.Where("id = ? AND user_id = ?", threadID, userID).Delete(&chatmodel.ChatThread{}).Error
}

// CheckThreadExists checks if a thread exists in the database
func (cs *chatStore) CheckThreadExists(threadID uuid.UUID) (bool, error) {
	var count int64
	err := cs.db.Model(&chatmodel.ChatThread{}).Where("id = ?", threadID).Count(&count).Error
	if err != nil {
		return false, err // Handle database errors
	}
	return count > 0, nil
}

// CheckThreadExistsAndBelongsToUser checks if the thread exists and belongs to the user
func (store *chatStore) CheckThreadExistsAndBelongsToUser(threadID, userID uuid.UUID) (bool, error) {
    var count int64
    // Adjust the query to check for both existence and ownership
    err := store.db.Model(&chatmodel.ChatThread{}).Where("id = ? AND user_id = ?", threadID, userID).Count(&count).Error
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

