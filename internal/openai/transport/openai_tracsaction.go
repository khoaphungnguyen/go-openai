package openaitransport

import (
	"context"
	"sync"

	"github.com/google/uuid"
	openaibusiness "github.com/khoaphungnguyen/go-openai/internal/openai/business"
)

// OpenAIHandler handles HTTP requests for OpenAI operations.

type OpenAIHandler struct {
    openAIService     *openaibusiness.OpenAIService
    ThreadSSEChannels map[uuid.UUID]chan string
    Mutex             *sync.RWMutex
    ctx               context.Context
    // CancelFuncs stores the cancel functions for each thread ID
    CancelFuncs map[uuid.UUID]context.CancelFunc
    CancelFuncsLLM map[uuid.UUID]context.CancelFunc
}

// NewOpenAIHandler creates a new instance of OpenAIHandler.
func NewOpenAIHandler(openAIService *openaibusiness.OpenAIService) *OpenAIHandler {
    ctx, _ := context.WithCancel(context.Background())
    return &OpenAIHandler{
        openAIService:     openAIService,
        ThreadSSEChannels: make(map[uuid.UUID]chan string),
        Mutex:             &sync.RWMutex{},
        ctx:               ctx,
        // Initialize the CancelFuncs map
        CancelFuncs: make(map[uuid.UUID]context.CancelFunc),
        CancelFuncsLLM: make(map[uuid.UUID]context.CancelFunc),
    }
}
