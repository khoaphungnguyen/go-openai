// chatbusiness contains the business logic for chat operations.
package chatbusiness

import chatstorage "github.com/khoaphungnguyen/go-openai/internal/chat/storage"

// ChatService provides methods for chat operations.
type ChatService struct {
	chatStore chatstorage.ChatStore
}

// NewChatService creates a new ChatService.
func NewChatService(chatStore chatstorage.ChatStore) *ChatService {
	return &ChatService{chatStore: chatStore}
}
