package storageuser

import (
	"errors"
	"time"

	"github.com/google/uuid"
	modeluser "github.com/khoaphungnguyen/go-openai/internal/user/model"
	"gorm.io/gorm"
)

// Create adds a new user to the database with a hashed password
func (store *userStore) Create(user *modeluser.User) error {
	// No need to hash here, as it should be done already in the business layer
	return store.db.Create(user).Error
}

// GetByEmail finds a user by email
func (store *userStore) GetUserByEmail(email string) (*modeluser.User, error) {
	var user modeluser.User
	err := store.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// Update modifies an existing user in the database
func (store *userStore) Update(user *modeluser.User) error {
	return store.db.Save(user).Error
}

// Delete removes a user from the database by UUID
func (store *userStore) Delete(email string) error {
	return store.db.Delete(&modeluser.User{}, email).Error
}

// GetByUUID finds a user by UUID
func (store *userStore) GetUserByUUID(id uuid.UUID) (*modeluser.User, error) {
	var user modeluser.User
	err := store.db.Where("id = ?", id).First(&user).Error
	return &user, err
}

// EmailVerified checks if the user's email has been verified
func (store *userStore) EmailVerified(email string) (bool, error) {
	var user modeluser.User
	if err := store.db.Select("email_verified").First(&user, email).Error; err != nil {
		return false, err
	}
	return user.EmailVerified, nil
}

// CheckLastLogin checks when the user last logged in and potentially performs some actions if necessary
func (store *userStore) CheckLastLogin(email string) (bool, error) {
	var user modeluser.User
	if err := store.db.First(&user, "email = ?", email).Error; err != nil {
		// Handle not found error separately if necessary
		if err == gorm.ErrRecordNotFound {
			return false, nil // or return a custom error indicating user not found
		}
		return false, err
	}

	// Define the duration after which a token should be refreshed (e.g., 30 days)
	tokenRefreshDuration := 30 * 24 * time.Hour

	// If LastLogin is nil, it means the user never logged in, handle this case as needed
	if user.LastLogin == nil {
		return false, errors.New("user never logged in") // or another appropriate action
	}

	// Check if the last login is within the token refresh duration
	if time.Since(*user.LastLogin) <= tokenRefreshDuration {
		// Last login is within the required duration, no need to refresh token
		return true, nil
	}

	// Last login is older than the required duration, token should be refreshed
	return false, nil
}
