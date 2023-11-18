package messagestorage

import (
	"fmt"

	"github.com/google/uuid"
	messagemodel "github.com/khoaphungnguyen/go-openai/internal/message/model"
	"gorm.io/gorm"
)

// CreateThread adds a new chat thread to the database.
func (ms *messageStore) CreateThread(thread *messagemodel.ChatThread) error {
	return ms.db.Create(thread).Error
}

// GetThreadByID retrieves a chat thread by its ID.
func (ms *messageStore) GetThreadByID(threadID uuid.UUID) (*messagemodel.ChatThread, error) {
	var thread messagemodel.ChatThread
	err := ms.db.First(&thread, "id = ?", threadID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Thread does not exist
		}
		return nil, fmt.Errorf("failed to retrieve thread: %w", err)
	}
	return &thread, nil
}

// GetThreadsByUserID retrieves all chat threads for a specific user
func (ms *messageStore) GetThreadsByUserID(userID uuid.UUID, limit, offset int) ([]messagemodel.ChatThread, error) {
	var threads []messagemodel.ChatThread
	err := ms.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&threads).Error
	return threads, err
}

// GetAllThreads retrieves all chat threads for a specific user.
func (ms *messageStore) GetAllThreads(userID uuid.UUID) ([]messagemodel.ChatThread, error) {
	return ms.GetThreadsByUserID(userID, -1, 0) // -1 for no limit
}

// CreateMessage adds a new message to a chat thread
func (ms *messageStore) CreateMessage(message *messagemodel.ChatMessage) error {
	return ms.db.Create(message).Error
}

// IsUserThreadOwner checks if a user is the owner of a specific thread.
func (ms *messageStore) IsUserThreadOwner(threadID, userID uuid.UUID) bool {
	var count int64
	ms.db.Model(&messagemodel.ChatThread{}).Where("id = ? AND user_id = ?", threadID, userID).Count(&count)
	return count > 0
}

// GetMessagesByThreadID retrieves messages of a chat thread with pagination.
func (ms *messageStore) GetMessagesByThreadID(threadID uuid.UUID, offset, limit int) ([]messagemodel.ChatMessage, error) {
	var messages []messagemodel.ChatMessage
	err := ms.db.Where("thread_id = ?", threadID).Offset(offset).Limit(limit).Find(&messages).Error
	return messages, err
}

// DeleteThread deletes a chat thread.
func (ms *messageStore) DeleteThread(threadID uuid.UUID, userID uuid.UUID) error {
	return ms.db.Where("id = ? AND user_id = ?", threadID, userID).Delete(&messagemodel.ChatThread{}).Error
}

// CheckThreadExists checks if a thread exists in the database.
func (ms *messageStore) CheckThreadExists(threadID uuid.UUID) (bool, error) {
	var count int64
	err := ms.db.Model(&messagemodel.ChatThread{}).Where("id = ?", threadID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check thread existence: %w", err)
	}
	return count > 0, nil
}

// CheckThreadExistsAndBelongsToUser checks if a thread exists and belongs to the user.
func (ms *messageStore) CheckThreadExistsAndBelongsToUser(threadID, userID uuid.UUID) (bool, error) {
	var count int64
	err := ms.db.Model(&messagemodel.ChatThread{}).Where("id = ? AND user_id = ?", threadID, userID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check thread ownership: %w", err)
	}
	return count > 0, nil
}
