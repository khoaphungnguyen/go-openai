// messagetransport handles HTTP requests and responses for chat operations.
package messagetransport

import messagebusiness "github.com/khoaphungnguyen/go-openai/internal/message/business"

// MessageHandler handles chat-related HTTP requests.
type MessageHandler struct {
	messsageService *messagebusiness.MessageService
}

// NewMessageHandler creates a new ChatHandler.
func NewMessageHandler(messsageService *messagebusiness.MessageService) *MessageHandler {
	return &MessageHandler{messsageService: messsageService}
}
