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

	// Store the last login time before updating it
	var lastLoginStr string
	if user.LastLogin != nil {
		lastLoginStr = user.LastLogin.Format(time.RFC3339)
	}

	// Update the last login time without affecting other fields
	now := time.Now()
	err = h.userService.UpdateLastLogin(user.ID, &now)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "Failed to update last login time"})
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
		c.JSON(500, gin.H{"Error": "Error Signing Token"})
		c.Abort()
		return
	}

	// Generate refresh token with user's UUID and full name
	signedRefreshToken, err := jwtWrapper.RefreshToken(user.ID.String(), user.FullName)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"Error": "Error Signing Refresh Token"})
		c.Abort()
		return
	}

	// Set the refresh token in an HTTP-only cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refreshToken",
		Value:    signedRefreshToken,
		HttpOnly: true,
		Path:     "/",
		Secure:   true, // Set to true if using HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(jwtWrapper.ExpirationHours * 3600),
	})

	// Return the access token and last login time in the JSON response
	c.JSON(200, gin.H{
		"token":     signedToken,
		"lastLogin": lastLoginStr,
	})
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

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		c.JSON(401, gin.H{
			"Error": "Token is expired",
		})
		c.Abort()
		return
	}

	// Parse UUID from the claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(500, gin.H{
			"Error": "Error parsing user ID",
		})
		c.Abort()
		return
	}

	isActive, err := h.userService.CheckLastLogin(userID)
	if err != nil {
		// Handle server errors separately
		c.JSON(500, gin.H{"Error": "Server error checking user status"})
		c.Abort()
		return
	}
	if !isActive {
		// User is not active, possibly due to being soft-deleted
		c.JSON(401, gin.H{"Error": "Inactive account. Restore to continue."})
		c.Abort()
		return
	}

	// Generate a new access token
	newAccessToken, err := jwtWrapper.GenerateToken(claims.UserID, claims.FullName)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"Error": "Error signing new token",
		})
		c.Abort()
		return
	}

	c.JSON(200, gin.H{
		"token": newAccessToken, // Return the new access token in the JSON response
	})
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

	// Retrieve the current user data
	user, err := h.userService.GetUserByUUID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	fmt.Println("email", user.Email, payload.Email)

	// Check if the new email already exists for another user
	if payload.Email != user.Email && h.userService.IsEmailExists(payload.Email, user.ID) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use."})
		return
	}

	// Update the user data with the payload
	user.FullName = payload.FullName
	user.Email = payload.Email

	// Update the user in the database
	if err := h.userService.UpdateUser(user, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteProfile handles the soft deletion of a user profile
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

	// Check if the user is already soft-deleted
	isSoftDeleted, err := h.userService.IsSoftDeleted(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check account's status"})
		return
	}

	if isSoftDeleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User account is already inactive"})
		return
	}

	if err := h.userService.SoftDeleteUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete your account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your account is deleted successfully"})
}

// RestoreProfile handles reactivating a soft-deleted user profile
func (h *UserHandler) RestoreProfile(c *gin.Context) {
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

	// Check if the user is soft-deleted
	isSoftDeleted, err := h.userService.IsSoftDeleted(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check account's status"})
		return
	}

	if !isSoftDeleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Your account is already active"})
		return
	}

	if err := h.userService.RestoreUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore your account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your account is restored successfully"})
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
