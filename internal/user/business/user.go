package businessuser

import storageuser "github.com/khoaphungnguyen/go-openai/internal/user/storage"

type UserService struct {
	userStore storageuser.UserStore
}

func NewUserService(userStore storageuser.UserStore) *UserService {
	return &UserService{userStore: userStore}
}
