// OpenAIStorage handles database operations for OpenAI interactions
package openaistorage

import (
	"github.com/google/uuid"
	openaimodel "github.com/khoaphungnguyen/go-openai/internal/openai/model"
	"gorm.io/gorm"
)

// OpenAIStore provides methods for OpenAI operations.
type OpenAIStore interface {
	CreateTransaction(transaction *openaimodel.OpenAITransaction) (uuid.UUID, error)
	GetTransactionsByUserID(userID uuid.UUID) ([]openaimodel.OpenAITransaction, error)
	GetTransactionsByThreadID(threadID uuid.UUID) ([]openaimodel.OpenAITransaction, error)
	GetTransactionByID(transactionID uuid.UUID) (*openaimodel.OpenAITransaction, error)
	UpdateTransaction(transaction *openaimodel.OpenAITransaction) error
	DeleteTransaction(id uuid.UUID) error
	CountUserTransactions(userID uuid.UUID) (int64, error)
	SummarizeUsage(userID uuid.UUID) (int64, error)
}

// openAIStore encapsulates the logic for storing and retrieving OpenAI data.
type openAIStore struct {
	db *gorm.DB
}

// NewOpenAIStore creates a new instance of OpenAIStorage
func NewOpenAIStore(db *gorm.DB) OpenAIStore {
	return &openAIStore{db: db}
}
