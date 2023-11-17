package chatbusiness

import (
	"errors"

	"github.com/google/uuid"
	chatmodel "github.com/khoaphungnguyen/go-openai/internal/chat/model"
)

// CreateThread creates a new chat thread
func (cs *ChatService) CreateThread(thread *chatmodel.ChatThread) error {
	return cs.chatStore.CreateThread(thread)
}

// GetThreadByID retrieves a chat thread by its ID
func (cs *ChatService) GetThreadByID(threadID uuid.UUID) (*chatmodel.ChatThread, error) {
	return cs.chatStore.GetThreadByID(threadID)
}

// GetThreadsByUserID retrieves all chat threads for a specific user
func (cs *ChatService) GetThreadsByUserID(userID uuid.UUID, limit, offset int) ([]chatmodel.ChatThread, error) {
	return cs.chatStore.GetThreadsByUserID(userID, limit, offset)
}

func (cs *ChatService) CreateMessage(userID uuid.UUID, message *chatmodel.ChatMessage) error {
    // First, check if the thread exists and belongs to the user
    threadExists, err := cs.chatStore.CheckThreadExistsAndBelongsToUser(message.ThreadID, userID)
    if err != nil {
        // handle error, perhaps log it and return an error to the handler
        return err
    }
    if !threadExists {
        // return an error that will be understood by your handler to mean forbidden action
        return errors.New("thread does not exist or you do not have permission to post in it")
    }

    // If checks pass, create the message
    return cs.chatStore.CreateMessage(message)
}


func (cs *ChatService) GetAllThreads(userID uuid.UUID) ([]chatmodel.ChatThread, error) {
	return cs.chatStore.GetAllThreads(userID)
}

func (cs *ChatService) GetMessagesByThreadID(threadID uuid.UUID, limit, offset int, userID uuid.UUID) ([]chatmodel.ChatMessage, error) {
	// Check if the user has access to the thread
	if !cs.chatStore.IsUserThreadOwner(threadID, userID) {
		return nil, errors.New("unauthorized access to thread")
	}

	return cs.chatStore.GetMessagesByThreadID(threadID, limit, offset)
}

func (cs *ChatService) DeleteThread(threadID uuid.UUID, userID uuid.UUID) error {
	// Check if the user has access to the thread
	if !cs.chatStore.IsUserThreadOwner(threadID, userID) {
		return errors.New("unauthorized access to thread")
	}

	return cs.chatStore.DeleteThread(threadID, userID)
}

// CheckThreadExists checks if a thread exists in the database
func (cs *ChatService) CheckThreadExists(threadID uuid.UUID) (bool, error) {
	return cs.chatStore.CheckThreadExists(threadID)
}


