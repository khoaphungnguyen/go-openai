package businessuser

import (
	"errors"
	"github.com/google/uuid"
	modeluser "github.com/khoaphungnguyen/go-openai/internal/user/model"
)

func (s *UserService) CreateUser(user *modeluser.User) error {
	// Hash the password before saving the user
	if err := user.HashPassword(user.Password); err != nil {
		return err
	}

	// Call the storage layer to create the user
	return s.userStore.Create(user)
}

func (s *UserService) UpdateUser(user *modeluser.User) error {
	// If the password is being updated, hash the new password
	if user.Password != "" {
		if err := user.HashPassword(user.Password); err != nil {
			return err
		}
	}

	// Call the storage layer to update the user
	return s.userStore.Update(user)
}

func (s *UserService) DeleteUser(email string) error {
	// Call the storage layer to delete the user
	return s.userStore.Delete(email)
}

func (s *UserService) GetUserByUUID(id uuid.UUID) (*modeluser.User, error) {
	// Call the storage layer to get the user by UUID
	return s.userStore.GetUserByUUID(id)
}

func (s *UserService) GetUserByEmail(email string) (*modeluser.User, error) {
	// Call the storage layer to get the user by email
	return s.userStore.GetUserByEmail(email)
}

func (s *UserService) VerifyUserPassword(email, password string) (bool, error) {
	user, err := s.GetUserByEmail(email)
	if err != nil {
		return false, err
	}

	// Check the provided password
	if user != nil {
		err = user.CheckPassword(password)
		if err != nil {
			// Password does not match
			return false, errors.New("incorrect password")
		}
		// Password matches
		return true, nil
	}

	return false, errors.New("user not found")
}
