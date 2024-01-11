package openaistorage

import (
	"github.com/google/uuid"
	openaimodel "github.com/khoaphungnguyen/go-openai/internal/openai/model"
)

// CreateTransaction saves a new transaction with the OpenAI API to the database.
func (s *openAIStore) CreateTransaction(transaction *openaimodel.OpenAITransaction) (uuid.UUID, error) {
    result := s.db.Create(transaction)
    if result.Error != nil {
        return uuid.Nil, result.Error
    }
    return transaction.ID, nil
}

// GetTransactionsByUserID finds OpenAI transactions for a given user ID.
func (s *openAIStore) GetTransactionsByUserID(userID uuid.UUID) ([]openaimodel.OpenAITransaction, error) {
	var transactions []openaimodel.OpenAITransaction
	err := s.db.Where("user_id = ?", userID).Find(&transactions).Error
	return transactions, err
}

// GetTransactionsByThreadID finds OpenAI transactions for a given thread ID.
func (s *openAIStore) GetTransactionsByThreadID(threadID uuid.UUID) ([]openaimodel.OpenAITransaction, error) {
	var transactions []openaimodel.OpenAITransaction
	err := s.db.Where("thread_id = ?", threadID).Find(&transactions).Error
	return transactions, err
}

// GetTransactionByID finds an OpenAI transaction by its ID.
func (s *openAIStore) GetTransactionByID(transactionID uuid.UUID) (*openaimodel.OpenAITransaction, error) {
    var transaction openaimodel.OpenAITransaction
    err := s.db.Where("id = ?", transactionID).First(&transaction).Error
    if err != nil {
        return nil, err
    }
    return &transaction, nil
}

// UpdateTransaction updates an existing OpenAI transaction.
func (s *openAIStore) UpdateTransaction(transaction *openaimodel.OpenAITransaction) error {
	return s.db.Save(transaction).Error
}

// DeleteTransaction removes a transaction record from the database.
func (s *openAIStore) DeleteTransaction(id uuid.UUID) error {
	return s.db.Delete(&openaimodel.OpenAITransaction{}, id).Error
}

// CountUserTransactions counts the total number of transactions made by a specific user.
func (s *openAIStore) CountUserTransactions(userID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&openaimodel.OpenAITransaction{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// SummarizeUsage calculates the total message length processed for a specific user.
func (s *openAIStore) SummarizeUsage(userID uuid.UUID) (int64, error) {
	var totalLength int64
	err := s.db.Model(&openaimodel.OpenAITransaction{}).Where("user_id = ?", userID).Select("sum(message_length)").Row().Scan(&totalLength)
	return totalLength, err
}
