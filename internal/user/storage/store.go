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
	GetAllUsers() ([]modeluser.User, error)
	Update(user *modeluser.User) error
	Delete(id uuid.UUID) error
	GetUserByUUID(id uuid.UUID) (*modeluser.User, error)
	CheckLastLogin(id uuid.UUID) (bool, error)
	EmailVerified(email string) (bool, error)
	UpdateLastLogin(userID uuid.UUID, lastLogin *time.Time) error
	IsEmailExists(email string) bool
	IsEmailExistsForOtherUser(email string, excludeUserID uuid.UUID) bool
	Restore(id uuid.UUID) error
	UpdateOmitFields(user *modeluser.User, omitFields ...string) error
	SoftDelete(id uuid.UUID) error
	IsSoftDeleted(userID uuid.UUID) (bool, error)
}

// userStore encapsulates the data storage logic for user operations.
type userStore struct {
	db *gorm.DB
}

// NewUserStore creates a new instance of a user store
func NewUserStore(db *gorm.DB) UserStore {
	return &userStore{db: db}
}
