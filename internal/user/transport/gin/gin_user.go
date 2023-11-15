package ginuser

import businessuser "github.com/khoaphungnguyen/go-openai/internal/user/business"

type UserHanlder struct {
	userService *businessuser.UserService
	JWTKey      string
}

func NewUserHandler(userService *businessuser.UserService, JWTKey string) *UserHanlder {
	return &UserHanlder{userService: userService, JWTKey: JWTKey}
}
