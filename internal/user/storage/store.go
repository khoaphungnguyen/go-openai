package userstorage

import (
	"time"

	"github.com/google/uuid"
	modeluser "github.com/khoaphungnguyen/go-openai/internal/user/model"
	"gorm.io/gorm"
)

// UserStore defines the interface for user storage operations
type UserStore interface {
	Create(user *modeluser.User) error
	GetUserByEmail(email string) (*modeluser.User, error)
	Update(user *modeluser.User) error
	Delete(id uuid.UUID) error
	GetUserByUUID(id uuid.UUID) (*modeluser.User, error)
	CheckLastLogin(id uuid.UUID) (bool, error)
	EmailVerified(email string) (bool, error)
	UpdateLastLogin(userID uuid.UUID, lastLogin *time.Time) error
	IsEmailExists(email string, excludeUserID uuid.UUID) bool
	SoftDelete(id uuid.UUID) error
	Restore(id uuid.UUID) error
}

// userStore implements the UserStore interface with gorm.DB
type userStore struct {
	db *gorm.DB
}

// NewUserStore creates a new instance of a user store
func NewUserStore(db *gorm.DB) UserStore {
	return &userStore{db: db}
}
