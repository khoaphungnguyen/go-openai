package notebusiness

import (
	"errors"

	"github.com/google/uuid"
	notemodel "github.com/khoaphungnguyen/go-openai/internal/note/model"
)

// CreateNote handles the creation of a new note.
func (ns *NoteService) CreateNote(note *notemodel.Note) error {
	if note == nil {
		return errors.New("note cannot be nil")
	}
	return ns.notestorage.CreateNote(note)
}

// GetNotesByUserID retrieves all notes for a specific user
func (ns *NoteService) GetNotesByUserID(userID uuid.UUID, limit, offset int) ([]*notemodel.Note, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}
	return ns.notestorage.GetNotesByUserID(userID, limit, offset)
}

// GetNoteByID retrieves a note by its ID.
func (ns *NoteService) GetNoteByID(userID, noteID uuid.UUID) (*notemodel.Note, error) {

	onwer := ns.notestorage.IsUserNoteOwner(noteID, userID)
	if !onwer {
		return nil, errors.New("user is not owner")
	}

	return ns.notestorage.GetNoteByID(noteID)
}

// DeleteNote deletes a note.
func (ns *NoteService) DeleteNote(userID, noteID uuid.UUID) error {
	onwer := ns.notestorage.IsUserNoteOwner(noteID, userID)
	if !onwer {
		return errors.New("user is not owner")
	}
	return ns.notestorage.DeleteNote(noteID, userID)
}

// UpdateNote updates a note.
func (ns *NoteService) UpdateNoteByID(userID, noteID uuid.UUID, note *notemodel.Note) error {
	onwer := ns.notestorage.IsUserNoteOwner(noteID, userID)
	if !onwer {
		return errors.New("user is not owner")
	}
	return ns.notestorage.UpdateNoteByID(noteID, note)
}
