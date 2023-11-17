// chatmodel defines the data structures used in the chat application.
package chatmodel

import (
	"time"

	"github.com/google/uuid"
)

// ChatMessage represents a single message in a chat thread.
type ChatMessage struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ThreadID  uuid.UUID `gorm:"type:uuid;index"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Role      string    `gorm:"type:varchar(50);not null"`
	Model     string    `gorm:"type:varchar(255)"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"default:now()"`
}

// TableName overrides the table name used by ChatMessage.
func (ChatMessage) TableName() string {
	return "chat_message"
}

// ChatThread represents a thread of chat messages.
type ChatThread struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Title     string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time `gorm:"default:now()"`
	UpdatedAt time.Time `gorm:"default:now()"`
}

// TableName overrides the table name used by ChatThread.
func (ChatThread) TableName() string {
	return "chat_thread"
}
