package notetransport

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	notemodel "github.com/khoaphungnguyen/go-openai/internal/note/model"
)

type NoteCreateRequest struct {
	Title string `json:"title"`
}

type NoteCreateResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type NoteResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Problem   string    `json:"problem"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NoteDetail struct {
	Problem   string `json:"problem"`
	Approach  string `json:"approach"`
	Solution  string `json:"solution"`
	ExtraNote string `json:"extra_note"`
}

type NoteDetailResponse struct {
	Title     string    `json:"title"`
	Problem   string    `json:"problem"`
	Approach  string    `json:"approach"`
	Solution  string    `json:"solution"`
	ExtraNote string    `json:"extra_note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateNote handles the creation of a new note.
func (nh *NoteHandler) CreateNote(c *gin.Context) {
	var payload NoteCreateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}
	note := &notemodel.Note{
		Title:  payload.Title,
		UserID: userID,
	}
	if err := nh.noteService.CreateNote(note); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to create note")
		return
	}

	respondWithJSON(c, http.StatusCreated, note)

}

// GetAllNoteByUserID handles the retrieval of all notes for a specific user.
func (nh *NoteHandler) GetAllNoteByUserID(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	notes, err := nh.noteService.GetNotesByUserID(userID, 20, 0)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to retrieve threads")
		return
	}

	var noteResponses []NoteResponse
	for _, note := range notes {
		noteResponses = append(noteResponses, NoteResponse{
			ID:        note.ID,
			Title:     note.Title,
			Problem:   note.Problem,
			CreatedAt: note.CreatedAt,
			UpdatedAt: note.UpdatedAt,
		})
	}

	respondWithJSON(c, http.StatusOK, noteResponses)
}

// GetNoteByID handles retrieving a single note by its ID.
func (nh *NoteHandler) GetNoteByID(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid note ID")
		return
	}

	note, err := nh.noteService.GetNoteByID(userID, noteID)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error retrieving note")
		return
	}

	// If note is not found, return a not found error instead of forbidden
	if note == nil {
		respondWithError(c, http.StatusNotFound, "Note not found")
		return
	}

	respondWithJSON(c, http.StatusOK, convertToNoteResponse(note))
}

// DeleteNote handles the deletion of a note.
func (nh *NoteHandler) DeleteNote(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid note ID")
		return
	}

	if err := nh.noteService.DeleteNote(userID, noteID); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	respondWithJSON(c, http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

// UpdateNote handles the updating of a note.
func (nh *NoteHandler) UpdateNote(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid note ID")
		return
	}

	var payload NoteDetail
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	note := &notemodel.Note{
		Problem:   payload.Problem,
		Approach:  payload.Approach,
		Solution:  payload.Solution,
		ExtraNote: payload.ExtraNote,
	}

	if err := nh.noteService.UpdateNoteByID(userID, noteID, note); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to update note")
		return
	}

	respondWithJSON(c, http.StatusOK, gin.H{"message": "Note updated successfully"})
}

// Helper functions
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, errors.New("userID not provided")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID")
	}

	return userID, nil
}

func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

func respondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

// convertToNoteResponse converts a note model to a note response.
func convertToNoteResponse(note *notemodel.Note) NoteDetailResponse {
	return NoteDetailResponse{
		Title:     note.Title,
		Problem:   note.Problem,
		Approach:  note.Approach,
		Solution:  note.Solution,
		ExtraNote: note.ExtraNote,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}
}
