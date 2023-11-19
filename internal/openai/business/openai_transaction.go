package openaibusiness

import (
	messagebusiness "github.com/khoaphungnguyen/go-openai/internal/message/business"
	openaistorage "github.com/khoaphungnguyen/go-openai/internal/openai/storage"
)

// OpenAIService provides business logic for OpenAI transactions.
type OpenAIService struct {
	openAIStore    openaistorage.OpenAIStore
	messageService *messagebusiness.MessageService // Reference to the message business service
}

// NewOpenAIService creates a new instance of OpenAIService.
func NewOpenAIService(openAIStore openaistorage.OpenAIStore, msgService *messagebusiness.MessageService) *OpenAIService {
	return &OpenAIService{
		openAIStore:    openAIStore,
		messageService: msgService,
	}
}
