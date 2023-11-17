package userstorage

import (
	"errors"
	"time"

	"github.com/google/uuid"
	modeluser "github.com/khoaphungnguyen/go-openai/internal/user/model"
	"gorm.io/gorm"
)

// ErrUserNotFound is the error returned when a user cannot be found.
var ErrUserNotFound = errors.New("user not found")

// Create adds a new user to the database, omitting the password field.
func (store *userStore) Create(user *modeluser.User) error {
	return store.db.Omit("password").Create(user).Error
}

// GetUserByEmail finds a user by email, including soft-deleted users.
func (store *userStore) GetUserByEmail(email string) (*modeluser.User, error) {
	var user modeluser.User
	result := store.db.Unscoped().Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, result.Error
}

// A soft-delete to remove a user from the database by UUID.
func (store *userStore) Delete(id uuid.UUID) error {
	return store.db.Delete(&modeluser.User{}, id).Error
}

// GetUserByUUID finds a user by UUID.
func (store *userStore) GetUserByUUID(id uuid.UUID) (*modeluser.User, error) {
	var user modeluser.User
	result := store.db.Where("id = ?", id).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, result.Error
}

// EmailVerified checks if a user's email has been verified.
func (store *userStore) EmailVerified(email string) (bool, error) {
	var user modeluser.User
	result := store.db.Select("email_verified").Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, ErrUserNotFound
	}
	return user.EmailVerified, result.Error
}

// CheckLastLogin checks if the user exists and has an active session.
func (store *userStore) CheckLastLogin(id uuid.UUID) (bool, error) {
	var user modeluser.User
	result := store.db.Where("id = ?", id).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, ErrUserNotFound
	}
	return result.Error == nil, result.Error
}

// UpdateLastLogin updates the last login time field in the database.
func (store *userStore) UpdateLastLogin(userID uuid.UUID, lastLogin *time.Time) error {
	return store.db.Model(&modeluser.User{}).Where("id = ?", userID).Update("last_login", lastLogin).Error
}

// IsEmailExists checks if the email exists in the system.
func (store *userStore) IsEmailExists(email string) bool {
	var count int64
	store.db.Model(&modeluser.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

// IsEmailExistsForOtherUser checks if the email exists for any user other than the one with the given UUID.
func (store *userStore) IsEmailExistsForOtherUser(email string, excludeUserID uuid.UUID) bool {
	var count int64
	store.db.Model(&modeluser.User{}).
		Where("email = ? AND id != ?", email, excludeUserID).
		Count(&count)
	return count > 0
}

// SoftDelete marks a user as deleted without actually removing them from the database.
func (store *userStore) SoftDelete(id uuid.UUID) error {
	var user modeluser.User
	// Find the user by ID
	result := store.db.First(&user, id)
	if result.Error != nil {
		return result.Error // Return any error (including user not found)
	}

	// Soft delete the user
	return store.db.Delete(&user).Error
}

// IsSoftDeleted checks if a user is soft-deleted.
func (store *userStore) IsSoftDeleted(userID uuid.UUID) (bool, error) {
	var user modeluser.User
	result := store.db.Unscoped().Select("id, deleted_at").Where("id = ?", userID).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil // User not found, not soft-deleted
	}
	if result.Error != nil {
		return false, result.Error // Error occurred during query
	}
	return user.DeletedAt != nil, nil // True if DeletedAt is not nil
}

// Restore reactivates a soft-deleted user by clearing the DeletedAt field.
func (store *userStore) Restore(id uuid.UUID) error {
	return store.db.Model(&modeluser.User{}).Unscoped().Where("id = ?", id).Update("deleted_at", gorm.Expr("NULL")).Error
}

// Update modifies an existing user in the database.
func (store *userStore) Update(user *modeluser.User) error {
	return store.db.Omit("password").Save(user).Error
}

// UpdateOmitFields modifies an existing user in the database while omitting specified fields.
func (store *userStore) UpdateOmitFields(user *modeluser.User, omitFields ...string) error {
	return store.db.Model(user).Omit(omitFields...).Save(user).Error
}
