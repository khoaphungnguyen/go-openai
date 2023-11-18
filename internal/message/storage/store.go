// messagestorage provides data persistence logic, specifically for message models.
package messagestorage

import (
	"github.com/google/uuid"
	messagemodel "github.com/khoaphungnguyen/go-openai/internal/message/model"
	"gorm.io/gorm"
)

// MessageStore provides methods for message operations.
type MessageStore interface {
	CreateThread(thread *messagemodel.ChatThread) error
	GetThreadByID(threadID uuid.UUID) (*messagemodel.ChatThread, error)
	GetThreadsByUserID(userID uuid.UUID, limit, offset int) ([]messagemodel.ChatThread, error)
	GetAllThreads(userID uuid.UUID) ([]messagemodel.ChatThread, error)
	CheckThreadExists(threadID uuid.UUID) (bool, error)
	IsUserThreadOwner(threadID, userID uuid.UUID) bool

	CreateMessage(message *messagemodel.ChatMessage) error
	GetMessagesByThreadID(threadID uuid.UUID, limit, offset int) ([]messagemodel.ChatMessage, error)
	DeleteThread(threadID uuid.UUID, userID uuid.UUID) error
	CheckThreadExistsAndBelongsToUser(threadID, userID uuid.UUID) (bool, error)
}

// messageStore encapsulates the logic for storing and retrieving message data.
type messageStore struct {
	db *gorm.DB
}

// NewMessageStore creates a new instance of messageStore.
func NewMessageStore(db *gorm.DB) MessageStore {
	return &messageStore{db: db}
}
