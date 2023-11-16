package userstorage

import (
	"errors"
	"time"

	"github.com/google/uuid"
	modeluser "github.com/khoaphungnguyen/go-openai/internal/user/model"
	"gorm.io/gorm"
)

// Create adds a new user to the database
func (store *userStore) Create(user *modeluser.User) error {
	// Omit the password field when saving to the database
	return store.db.Omit("password").Create(user).Error
}

// GetUserByEmail finds a user by email
func (store *userStore) GetUserByEmail(email string) (*modeluser.User, error) {
	var user modeluser.User
	err := store.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update modifies an existing user in the database
func (store *userStore) Update(user *modeluser.User) error {
	return store.db.Omit("password").Save(user).Error
}

// Delete removes a user from the database by UUID
func (store *userStore) Delete(id uuid.UUID) error {
	return store.db.Delete(&modeluser.User{}, id).Error
}

// GetUserByUUID finds a user by UUID
func (store *userStore) GetUserByUUID(id uuid.UUID) (*modeluser.User, error) {
	var user modeluser.User
	err := store.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// EmailVerified checks if the user's email has been verified
func (store *userStore) EmailVerified(email string) (bool, error) {
	var user modeluser.User
	err := store.db.Select("email_verified").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("user not found")
		}
		return false, err
	}
	return user.EmailVerified, nil
}

// CheckLastLogin checks if the user exists and has an active session
func (store *userStore) CheckLastLogin(id uuid.UUID) (bool, error) {
	var user modeluser.User
	err := store.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User not found or soft-deleted, treat as inactive
			return false, nil
		}
		// Other errors are treated as server errors
		return false, err
	}
	// User exists and is active
	return true, nil
}

// UpdateLastLogin updates the last login time field in the database
func (store *userStore) UpdateLastLogin(userID uuid.UUID, lastLogin *time.Time) error {
	return store.db.Model(&modeluser.User{}).Where("id = ?", userID).Update("last_login", lastLogin).Error
}

// IsEmailExists checks if the email exists for any user other than the one with the given UUID
func (store *userStore) IsEmailExists(email string, excludeUserID uuid.UUID) bool {
	var count int64
	store.db.Model(&modeluser.User{}).Where("email = ? AND id != ?", email, excludeUserID).Count(&count)
	return count > 0
}

func (store *userStore) SoftDelete(id uuid.UUID) error {
	return store.db.Model(&modeluser.User{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// Restore reactivates a soft-deleted user by clearing the DeletedAt field
func (store *userStore) Restore(id uuid.UUID) error {
	return store.db.Model(&modeluser.User{}).Unscoped().Where("id = ?", id).Update("deleted_at", gorm.Expr("NULL")).Error
}
