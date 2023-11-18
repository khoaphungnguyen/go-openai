// openaimodel defines the data structures used for OpenAI interactions.
package openaimodel

import (
	"time"

	"github.com/google/uuid"
)

// OpenAITransaction represents a record of an interaction with the OpenAI API.
type OpenAITransaction struct {
	ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID        uuid.UUID `gorm:"type:uuid"`
	ThreadID      uuid.UUID `gorm:"type:uuid;index"`
	MessageID     uuid.UUID `gorm:"type:uuid"`
	Model         string    `gorm:"type:varchar(255)"`
	Role          string    `gorm:"type:varchar(50);not null"`
	MessageLength int       `gorm:"type:int"` // Tracks the length of the user's message
	ProcessTime   time.Time `gorm:"default:now()"`
}

// TableName overrides the table name used by OpenAITransaction.
func (OpenAITransaction) TableName() string {
	return "openai_transaction"
}

// OpenAITransactionInput represents the input data for creating a new OpenAI transaction.
type OpenAITransactionInput struct {
	UserID  string `json:"userID"`
	Message string `json:"message"`
	Model   string `json:"model"`
	Role    string `json:"role"`
}
