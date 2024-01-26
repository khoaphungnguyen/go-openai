// notetransport handles HTTP requests and responses for note operations.
package notetransport

import (
	notebusiness "github.com/khoaphungnguyen/go-openai/internal/note/business"
)

// NoteHandler handles note-related HTTP requests.
type NoteHandler struct {
	noteService *notebusiness.NoteService
}

// NewMessageHandler creates a new ChatHandler.
func NewNoteHandler(noteService *notebusiness.NoteService) *NoteHandler {
	return &NoteHandler{noteService: noteService}
}
