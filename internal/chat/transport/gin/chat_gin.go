// chatgin handles HTTP requests and responses for chat operations.
package chatgin

import chatbusiness "github.com/khoaphungnguyen/go-openai/internal/chat/business"

// ChatHandler handles chat-related HTTP requests.
type ChatHandler struct {
	chatService *chatbusiness.ChatService
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(chatService *chatbusiness.ChatService) *ChatHandler {
    return &ChatHandler{chatService: chatService}
}