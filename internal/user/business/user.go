package userbusiness

import storageuser "github.com/khoaphungnguyen/go-openai/internal/user/storage"

// UserService provides business logic for user operations.
type UserService struct {
	userStore storageuser.UserStore
}

// NewUserService creates a new instance of UserService.
func NewUserService(userStore storageuser.UserStore) *UserService {
	return &UserService{userStore: userStore}
}
