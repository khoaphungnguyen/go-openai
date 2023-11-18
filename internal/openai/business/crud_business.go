package openaibusiness

import (
	"github.com/google/uuid"
	openaimodel "github.com/khoaphungnguyen/go-openai/internal/openai/model"
)

// ProcessNewTransaction handles the processing of a new OpenAI transaction.
func (s *OpenAIService) CreateTransaction(userID uuid.UUID, inputData string, model string, role string) error {
	// Simulate OpenAI API interaction and create a transaction record
	transaction := &openaimodel.OpenAITransaction{
		UserID: userID,
		// Assume MessageID and ThreadID are set appropriately
		Model:         model,
		Role:          role,
		MessageLength: len(inputData),
		// ProcessTime is automatically set by GORM
	}

	return s.openAIStore.CreateTransaction(transaction)
}

// UpdateTransaction updates an existing OpenAI transaction.
func (s *OpenAIService) UpdateTransaction(transaction *openaimodel.OpenAITransaction) error {
	return s.openAIStore.UpdateTransaction(transaction)
}

// DeleteTransaction deletes a transaction by its ID.
func (s *OpenAIService) DeleteTransaction(transactionID uuid.UUID) error {
	return s.openAIStore.DeleteTransaction(transactionID)
}

// GetTransactionByID retrieves a specific transaction by its ID.
func (s *OpenAIService) GetTransactionsByThreadID(threadID uuid.UUID) ([]openaimodel.OpenAITransaction, error) {
	return s.openAIStore.GetTransactionsByThreadID(threadID)
}

// GetTransactionsByUserID retrieves all transactions for a specific user.
func (s *OpenAIService) GetTransactionsByUserID(userID uuid.UUID) ([]openaimodel.OpenAITransaction, error) {
	return s.openAIStore.GetTransactionsByUserID(userID)
}

// GetTransactionsByTran retrieves all transactions for a specific user.
func (s *OpenAIService) GetTransactionByID(transactionID uuid.UUID) (*openaimodel.OpenAITransaction, error) {
	return s.openAIStore.GetTransactionByID(transactionID)
}

// CountUserTransactions counts the total number of transactions for a specific user.
func (s *OpenAIService) CountUserTransactions(userID uuid.UUID) (int64, error) {
	return s.openAIStore.CountUserTransactions(userID)
}

// SummarizeUserUsage calculates the total message length processed for a specific user.
func (s *OpenAIService) SummarizeUserUsage(userID uuid.UUID) (int64, error) {
	return s.openAIStore.SummarizeUsage(userID)
}
