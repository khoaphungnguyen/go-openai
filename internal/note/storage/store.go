// notestorage provides data persistence logic, specifically for note models.
package notestorage

import (
	"github.com/google/uuid"
	notemodel "github.com/khoaphungnguyen/go-openai/internal/note/model"
	"gorm.io/gorm"
)

// NoteStore provides methods for note operations.
type NoteStore interface {
	CreateNote(note *notemodel.Note) error
	GetNoteByID(noteID uuid.UUID) (*notemodel.Note, error)
	GetNotesByUserID(userID uuid.UUID, limit, offset int) ([]*notemodel.Note, error)
	GetAllNotes(userID uuid.UUID) ([]*notemodel.Note, error)
	CheckNoteExists(noteID uuid.UUID) (bool, error)
	IsUserNoteOwner(noteID, userID uuid.UUID) bool
	DeleteNote(noteID uuid.UUID, userID uuid.UUID) error
	CheckNoteExistsAndBelongsToUser(noteID, userID uuid.UUID) (bool, error)
	UpdateNoteByID(noteID uuid.UUID, note *notemodel.Note) error
}

// noteStore encapsulates the logic for storing and retrieving note data.
type noteStore struct {
	db *gorm.DB
}

// NewNoteStore creates a new instance of noteStore.
func NewNoteStore(db *gorm.DB) NoteStore {
	return &noteStore{db: db}
}
