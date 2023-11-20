package openaitransport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khoaphungnguyen/go-openai/internal/common"
	openaimodel "github.com/khoaphungnguyen/go-openai/internal/openai/model"
)

// CreateTransaction handles the creation of a new OpenAI transaction (HTTP Handler).
func (h *OpenAIHandler) CreateTransaction(c *gin.Context) {
	userID, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	var inputData openaimodel.OpenAITransactionInput
	if err := c.ShouldBindJSON(&inputData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err = h.createTransaction(userID, inputData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction created"})
}



// Core logic for creating a transaction
func (h *OpenAIHandler) createTransaction(userID uuid.UUID, inputData openaimodel.OpenAITransactionInput) error {
	threadID, err := uuid.Parse(inputData.ThreadID)
	if err != nil {
		return err
	}

	return h.openAIService.CreateTransaction(userID, threadID, inputData.Message, inputData.Model, inputData.Role)
}

// GetTransactionsByUserID handles fetching transactions for a specific user.
func (h *OpenAIHandler) GetTransactionsByUserID(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	transactions, err := h.openAIService.GetTransactionsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// UpdateTransaction handles the updating of an existing OpenAI transaction.
func (h *OpenAIHandler) UpdateTransaction(c *gin.Context) {
	var transaction openaimodel.OpenAITransaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.openAIService.UpdateTransaction(&transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction updated"})
}

// DeleteTransaction handles the deletion of an OpenAI transaction.
func (h *OpenAIHandler) DeleteTransaction(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("transactionID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	if err := h.openAIService.DeleteTransaction(transactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted"})
}

// GetTransactionByID handles fetching a specific transaction by its ID.
func (h *OpenAIHandler) GetTransactionByID(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("transactionID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	transaction, err := h.openAIService.GetTransactionByID(transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}
