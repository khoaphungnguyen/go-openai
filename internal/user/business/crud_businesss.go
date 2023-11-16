package userbusiness

import (
	"errors"

	"github.com/google/uuid"
	modeluser "github.com/khoaphungnguyen/go-openai/internal/user/model"
	"github.com/khoaphungnguyen/go-openai/internal/user/utils"
)

func (s *UserService) CreateUser(user *modeluser.User) error {
	// Hash the password
	hashedPassword, salt, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	// Store the hashed password and salt in the user struct
	user.PasswordHash = hashedPassword
	user.Salt = salt

	// Save the user to the database
	return s.userStore.Create(user)
}

func (s *UserService) UpdateUser(user *modeluser.User) error {
	if user.PasswordHash != "" {
		hashedPassword, salt, err := utils.HashPassword(user.PasswordHash)
		if err != nil {
			return err
		}
		user.PasswordHash = hashedPassword
		user.Salt = salt
	}

	return s.userStore.Update(user)
}

func (s *UserService) DeleteUser(id uuid.UUID) error {
	return s.userStore.Delete(id)
}

func (s *UserService) GetUserByUUID(id uuid.UUID) (*modeluser.User, error) {
	return s.userStore.GetUserByUUID(id)
}

func (s *UserService) GetUserByEmail(email string) (*modeluser.User, error) {
	return s.userStore.GetUserByEmail(email)
}

func (s *UserService) VerifyUserPassword(email, password string) (bool, error) {
	user, err := s.GetUserByEmail(email)
	if err != nil {
		return false, err
	}

	if user != nil {
		err = utils.CheckPassword(user.PasswordHash, user.Salt, password)
		if err != nil {
			return false, errors.New("incorrect password")
		}
		return true, nil
	}

	return false, errors.New("user not found")
}

// IsEmailExists checks if the provided email exists for any user other than the one with the given UUID
func (s *UserService) IsEmailExists(email string, excludeUserID uuid.UUID) bool {
	return s.userStore.IsEmailExists(email, excludeUserID)
}

// SoftDeleteUser marks a user as deleted without actually removing them from the database
func (s *UserService) SoftDeleteUser(id uuid.UUID) error {
	return s.userStore.SoftDelete(id)
}

// RestoreUser reactivates a soft-deleted user
func (s *UserService) RestoreUser(id uuid.UUID) error {
	return s.userStore.Restore(id)
}
