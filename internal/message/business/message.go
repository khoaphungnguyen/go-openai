// messagebusiness contains the business logic for message operations.
package messagebusiness

import messagestorage "github.com/khoaphungnguyen/go-openai/internal/message/storage"

// MessageService provides methods for message operations.
type MessageService struct {
	messageStore messagestorage.MessageStore
}

// NewMessageService creates a new MessageService.
func NewMessageService(messageStore messagestorage.MessageStore) *MessageService {
	return &MessageService{messageStore: messageStore}
}
