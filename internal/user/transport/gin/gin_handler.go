package usergin

import (
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

	// Generate JWT token with user's UUID and full name
	signedToken, err := jwtWrapper.GenerateToken(user.ID.String(), user.FullName)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"Error": "Error Signing Token",
		})
		c.Abort()
		return
	}

	// Generate refresh token with user's UUID and full name
	signedRefreshToken, err := jwtWrapper.RefreshToken(user.ID.String(), user.FullName)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"Error": "Error Signing Refresh Token",
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
	var payload modeluser.UserUpdatePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not provided"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Retrieve the existing user data
	user, err := h.userService.GetUserByUUID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update the user data with the payload
	user.FullName = payload.FullName
	user.Email = payload.Email

	// Update the user in the database
	if err := h.userService.UpdateUser(user); err != nil { // No '&' before user
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (h *UserHandler) DeleteProfile(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not provided"})
		return
	}

	userIDStr, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID is not valid"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
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
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByUUID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	publicUser := user.ToPublicUser()
	c.JSON(http.StatusOK, publicUser)
}

// Renew token from the refresh token
func (h *UserHandler) RenewAccessToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refreshToken")
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

	claims, err := jwtWrapper.ValidateToken(refreshToken)
	if err != nil {
		c.JSON(401, gin.H{
			"Error": "Invalid Token",
		})
		c.Abort()
		return
	}

	// Ensure the token is not expired
	if claims.ExpiresAt < time.Now().UTC().Unix() {
		c.JSON(401, gin.H{
			"Error": "Token is expired",
		})
		c.Abort()
		return
	}

	// Use the claims from the refresh token to generate a new access token
	newAccessToken, err := jwtWrapper.GenerateToken(claims.UserID, claims.FullName)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"Error": "Error Signing Token",
		})
		c.Abort()
		return
	}

	c.JSON(200, gin.H{
		"token": newAccessToken,
	})
}
