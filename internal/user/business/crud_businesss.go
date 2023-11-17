package userbusiness

import (
	"errors"
	"time"

	"github.com/google/uuid"
	modeluser "github.com/khoaphungnguyen/go-openai/internal/user/model"
	"github.com/khoaphungnguyen/go-openai/internal/user/utils"
)

// CreateUser handles the creation of a new user, including password hashing.
func (s *UserService) CreateUser(fullName, email, password string, role modeluser.Role) error {
	hashedPassword, salt, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	user := &modeluser.User{
		FullName:     fullName,
		Email:        email,
		PasswordHash: hashedPassword,
		Salt:         salt,
		Role:         role,
	}

	return s.userStore.Create(user)
}

// UpdateUser updates an existing user's information.
func (s *UserService) UpdateUser(userID uuid.UUID, fullName, email string) error {
	user, err := s.userStore.GetUserByUUID(userID)
	if err != nil {
		return err
	}

	user.FullName = fullName
	user.Email = email

	return s.userStore.Update(user)
}

// UpdateLastLogin updates only the last login time of the user
func (s *UserService) UpdateLastLogin(userID uuid.UUID, lastLogin *time.Time) error {
	return s.userStore.UpdateLastLogin(userID, lastLogin)
}

// DeleteUser deletes a user by UUID.
func (s *UserService) DeleteUser(id uuid.UUID) error {
	return s.userStore.Delete(id)
}

// GetUserByUUID retrieves a user by their UUID.
func (s *UserService) GetUserByUUID(id uuid.UUID) (*modeluser.User, error) {
	return s.userStore.GetUserByUUID(id)
}

// VerifyUserPassword checks if the provided password matches the user's stored password.
func (s *UserService) VerifyUserPassword(email, password string) (bool, error) {
	user, err := s.userStore.GetUserByEmail(email)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, errors.New("user not found")
	}

	return utils.CheckPassword(user.PasswordHash, user.Salt, password) == nil, nil
}

// SoftDeleteUser marks a user as deleted without actually removing them from the database.
func (s *UserService) SoftDeleteUser(id uuid.UUID) error {
	return s.userStore.SoftDelete(id)
}

// RestoreUser reactivates a soft-deleted user.
func (s *UserService) RestoreUser(id uuid.UUID) error {
	return s.userStore.Restore(id)
}

// IsSoftDeleted checks if a user is soft-deleted.
func (s *UserService) IsSoftDeleted(userID uuid.UUID) (bool, error) {
	return s.userStore.IsSoftDeleted(userID)
}

func (s *UserService) GetUserByEmail(email string) (*modeluser.User, error) {
	return s.userStore.GetUserByEmail(email)
}

// IsEmailExists checks if the provided email exists in the system.
func (s *UserService) IsEmailExists(email string) bool {
    return s.userStore.IsEmailExists(email)
}

// IsEmailExistsForOtherUser checks if the provided email exists for any user other than the one with the given UUID.
func (s *UserService) IsEmailExistsForOtherUser(email string, excludeUserID uuid.UUID) bool {
    return s.userStore.IsEmailExistsForOtherUser(email, excludeUserID)
}

// CheckLastLogin checks if the user exists and has an active session.
func (s *UserService) CheckLastLogin(userID uuid.UUID) (bool, error) {
	return s.userStore.CheckLastLogin(userID)
}
