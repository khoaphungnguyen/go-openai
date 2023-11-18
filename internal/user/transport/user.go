package usertransport

import businessuser "github.com/khoaphungnguyen/go-openai/internal/user/business"

type UserHandler struct {
	userService *businessuser.UserService
	JWTKey      string
}

func NewUserHandler(userService *businessuser.UserService, JWTKey string) *UserHandler {
	return &UserHandler{userService: userService, JWTKey: JWTKey}
}
