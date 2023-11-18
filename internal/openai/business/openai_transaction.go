package openaibusiness

import openaistorage "github.com/khoaphungnguyen/go-openai/internal/openai/storage"

// OpenAIService provides business logic for OpenAI transactions.
type OpenAIService struct {
	openAIStore openaistorage.OpenAIStore
}

// NewOpenAIService creates a new instance of OpenAIService.
func NewOpenAIService(openAIStore openaistorage.OpenAIStore) *OpenAIService {
	return &OpenAIService{
		openAIStore: openAIStore,
	}
}
