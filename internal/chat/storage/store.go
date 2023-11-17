// chatstorage provides data persistence logic, specifically for chat models.
package chatstorage

import (
	"github.com/google/uuid"
	chatmodel "github.com/khoaphungnguyen/go-openai/internal/chat/model"
	"gorm.io/gorm"
)


// ChatStore provides methods for chat operations.
type ChatStore interface {
	CreateThread(thread *chatmodel.ChatThread) error
	GetThreadByID(threadID uuid.UUID) (*chatmodel.ChatThread, error)
	GetThreadsByUserID(userID uuid.UUID, limit, offset int) ([]chatmodel.ChatThread, error)
	GetAllThreads(userID uuid.UUID) ([]chatmodel.ChatThread, error)
	CheckThreadExists(threadID uuid.UUID) (bool, error)
	IsUserThreadOwner(threadID, userID uuid.UUID) bool

	CreateMessage(message *chatmodel.ChatMessage) error
	GetMessagesByThreadID(threadID uuid.UUID, limit, offset int) ([]chatmodel.ChatMessage, error)
	DeleteThread(threadID uuid.UUID, userID uuid.UUID) error
	CheckThreadExistsAndBelongsToUser(threadID, userID uuid.UUID) (bool, error)
}

// chatStore encapsulates the logic for storing and retrieving chat data.
type chatStore struct {
	db *gorm.DB
}

// NewChatStore creates a new instance of chatStore.
func NewChatStore(db *gorm.DB) ChatStore {
	return &chatStore{db: db}
}
