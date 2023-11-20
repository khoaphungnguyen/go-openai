package openaibusiness

import (
	"github.com/google/uuid"
	messagemodel "github.com/khoaphungnguyen/go-openai/internal/message/model"
	openaimodel "github.com/khoaphungnguyen/go-openai/internal/openai/model"
)

func (s *OpenAIService) CreateTransaction(userID, threadID uuid.UUID, message, model, role string) error {
	chatMessage := &messagemodel.ChatMessage{
		ThreadID: threadID,
		UserID:   userID,
		Content:  message,
		Role:     role,
	}

	// Save the message using the message service
	if err := s.messageService.CreateMessage(userID, chatMessage); err != nil {
		return err
	}

	// Create and save the OpenAI transaction record
	transaction := &openaimodel.OpenAITransaction{
		UserID:        userID,
		ThreadID:      threadID,
		MessageID:     chatMessage.ID,
		Model:         model,
		Role:          role,
		MessageLength: len(message),
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
