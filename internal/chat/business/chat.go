package chatbusiness

import chatstorage "github.com/khoaphungnguyen/go-openai/internal/chat/storage"

type ChatService struct {
	chatStore chatstorage.ChatStore
}

func NewChatService(chatStore chatstorage.ChatStore) *ChatService {
	return &ChatService{chatStore: chatStore}
}