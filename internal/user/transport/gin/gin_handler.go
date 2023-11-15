package usergin

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	userauth "github.com/khoaphungnguyen/go-openai/internal/user/auth"
	modeluser "github.com/khoaphungnguyen/go-openai/internal/user/model"
	"github.com/khoaphungnguyen/go-openai/internal/user/utils"
)

// LoginPayload login body
type LoginPayload struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse token response
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshtoken"`
}

// Signup handles new user registration
func (h *UserHandler) Signup(c *gin.Context) {
	var user modeluser.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	fmt.Println(&user)
	if err := h.userService.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Inputs"})
		return
	}

	user, err := h.userService.GetUserByEmail(payload.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Email or Password"})
		return
	}

	if err := utils.CheckPassword(user.PasswordHash, user.Salt, payload.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Email or Password"})
		return
	}
	jwtWrapper := userauth.JwtWrapper{
		SecretKey:         h.JWTKey,
		Issuer:            "AuthService",
		ExpirationMinutes: 30,
		ExpirationHours:   12,
	}
	signedToken, err := jwtWrapper.GenerateToken(user.Email)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"Error": "Error Signing Token",
		})
		c.Abort()
		return
	}
	signedRefreshToken, err := jwtWrapper.RefreshToken(user.Email)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"Error": "Error Signing Token",
		})
		c.Abort()
		return
	}
	token := LoginResponse{
		Token:        signedToken,
		RefreshToken: signedRefreshToken,
	}
	c.JSON(200, token)
}

// UpdateProfile handles updating user information
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var user modeluser.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	userID, _ := c.Get("userID") // Retrieve the user ID from the context

	// Assume you have a method to convert the userID to the correct type
	// and that your UpdateUser method accepts a user model with ID set
	user.ID = userID.(uuid.UUID)

	if err := h.userService.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteProfile deletes a user profile
func (h *UserHandler) DeleteProfile(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not provided"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID is not valid"})
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// Profile retrieves the user's profile information
func (h *UserHandler) Profile(c *gin.Context) {
	userID, _ := c.Get("userID") // Retrieve the user ID from the context

	user, err := h.userService.GetUserByUUID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Renew token from the refresh token
func (h *UserHandler) RenewAccessToken(c *gin.Context) {
	token, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid Inputs",
		})
		c.Abort()
		return
	}
	jwtWrapper := userauth.JwtWrapper{
		SecretKey:         h.JWTKey,
		Issuer:            "AuthService",
		ExpirationMinutes: 30,
		ExpirationHours:   12,
	}
	claims, err := jwtWrapper.ValidateToken(token)
	if err != nil {
		c.JSON(401, gin.H{
			"Error": "Invalid Token",
		})
		c.Abort()
		return
	}
	if claims.ExpiresAt < time.Now().Add(time.Minute*30).Unix() {
		c.JSON(401, gin.H{
			"Error": "Token is expired",
		})
		c.Abort()
		return
	}
	// convert id to int
	email := claims.Audience
	if err != nil {
		c.JSON(401, gin.H{
			"Error": "Invalid Token",
		})
		c.Abort()
		return
	}
	signedToken, err := jwtWrapper.GenerateToken(email)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"Error": "Error Signing Token",
		})
		c.Abort()
		return
	}
	token = signedToken
	c.JSON(200, gin.H{
		"token": token,
	})

}
