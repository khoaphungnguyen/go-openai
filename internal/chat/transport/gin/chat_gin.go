package chatgin

import chatbusiness "github.com/khoaphungnguyen/go-openai/internal/chat/business"

type ChatHandler struct {
	chatService *chatbusiness.ChatService
}

func NewChatHandler(chatService *chatbusiness.ChatService) *ChatHandler {
    return &ChatHandler{chatService: chatService}
}