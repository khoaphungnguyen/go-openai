package chatmodel

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ThreadID  uuid.UUID `gorm:"type:uuid;index"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Role      string    `gorm:"type:varchar(50);not null"`
	Model     string    `gorm:"type:varchar(255)"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"default:now()"`
}

func (ChatMessage) TableName() string {
	return "chat_message"
}

type ChatThread struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Title     string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time `gorm:"default:now()"`
	UpdatedAt time.Time `gorm:"default:now()"`
}

func (ChatThread) TableName() string {
	return "chat_thread"
}
