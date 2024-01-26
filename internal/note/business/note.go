// notebusiness contains the business logic for message operations.
package notebusiness

import (
	notestorage "github.com/khoaphungnguyen/go-openai/internal/note/storage"
)

// NoteService provides methods for message operations.
type NoteService struct {
	notestorage notestorage.NoteStore
}

// NewNoteService creates a new NoteService.
func NewNoteService(notestorage notestorage.NoteStore) *NoteService {
	return &NoteService{notestorage: notestorage}
}
