package openaitransport

import openaibusiness "github.com/khoaphungnguyen/go-openai/internal/openai/business"

// OpenAIHandler handles HTTP requests for OpenAI operations.
type OpenAIHandler struct {
    openAIService *openaibusiness.OpenAIService
}

// NewOpenAIHandler creates a new instance of OpenAIHandler.
func NewOpenAIHandler(openAIService *openaibusiness.OpenAIService) *OpenAIHandler {
    return &OpenAIHandler{
        openAIService: openAIService,
    }
}