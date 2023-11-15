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
	return store.db.Save(user).Error
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

// CheckLastLogin checks when the user last logged in
func (store *userStore) CheckLastLogin(id uuid.UUID) (bool, error) {
	var user modeluser.User
	err := store.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("user not found")
		}
		return false, err
	}

	tokenRefreshDuration := 30 * 24 * time.Hour
	if user.LastLogin == nil || time.Since(*user.LastLogin) > tokenRefreshDuration {
		return false, nil // Token should be refreshed or user never logged in
	}
	return true, nil // Last login is within the required duration
}
