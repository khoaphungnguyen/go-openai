package notestorage

import (
	"fmt"

	"github.com/google/uuid"
	notemodel "github.com/khoaphungnguyen/go-openai/internal/note/model"
	"gorm.io/gorm"
)

// CreateNote adds a new note to the database.
func (ns *noteStore) CreateNote(note *notemodel.Note) error {
	return ns.db.Create(note).Error
}

// GetAllNotes retrieves all notes for a specific user.
func (ns *noteStore) GetAllNotes(userID uuid.UUID) ([]*notemodel.Note, error) {
	var notes []*notemodel.Note
	err := ns.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&notes).Error
	return notes, err
}

// GetNotesByUserID retrieves all notes for a specific user
func (ns *noteStore) GetNotesByUserID(userID uuid.UUID, limit, offset int) ([]*notemodel.Note, error) {
	var notes []*notemodel.Note
	err := ns.db.Where("user_id = ?", userID).Order("updated_at DESC").Limit(limit).Offset(offset).Find(&notes).Error
	return notes, err
}

// GetNoteByID retrieves a note by its ID.
func (ns *noteStore) GetNoteByID(noteID uuid.UUID) (*notemodel.Note, error) {
	var note notemodel.Note
	err := ns.db.First(&note, "id = ?", noteID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Note does not exist
		}
		return nil, fmt.Errorf("failed to retrieve note: %w", err)
	}
	return &note, nil
}

// IsUserNoteOwner checks if a user is the owner of a specific note.
func (ns *noteStore) IsUserNoteOwner(noteID, userID uuid.UUID) bool {
	var count int64
	ns.db.Model(&notemodel.Note{}).Where("id = ? AND user_id = ?", noteID, userID).Count(&count)
	return count > 0
}

// DeleteNote deletes a note.
func (ns *noteStore) DeleteNote(noteID uuid.UUID, userID uuid.UUID) error {
	return ns.db.Where("id = ? AND user_id = ?", noteID, userID).Delete(&notemodel.Note{}).Error
}

// CheckNoteExists checks if a note exists in the database.
func (ns *noteStore) CheckNoteExists(noteID uuid.UUID) (bool, error) {
	var count int64
	err := ns.db.Model(&notemodel.Note{}).Where("id = ?", noteID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check note existence: %w", err)
	}
	return count > 0, nil
}

// CheckNoteExistsAndBelongsToUser checks if a note exists and belongs to the user.
func (ns *noteStore) CheckNoteExistsAndBelongsToUser(noteID, userID uuid.UUID) (bool, error) {
	var count int64
	err := ns.db.Model(&notemodel.Note{}).Where("id = ? AND user_id = ?", noteID, userID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check note ownership: %w", err)
	}
	return count > 0, nil
}

// UpdateNote updates a note in the database.
func (ns *noteStore) UpdateNoteByID(noteID uuid.UUID, note *notemodel.Note) error {
	return ns.db.Model(&notemodel.Note{}).Where("id = ?", noteID).Updates(note).Error
}
