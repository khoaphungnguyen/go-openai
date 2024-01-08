package usertransport

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	userauth "github.com/khoaphungnguyen/go-openai/internal/user/auth"
	modeluser "github.com/khoaphungnguyen/go-openai/internal/user/model"
)

type UserRegistration struct {
	FullName string `json:"fullName" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}

// LoginPayload login body
type LoginPayload struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserUpdatePayload struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

// Signup handles new user registration.
func (h *UserHandler) Signup(c *gin.Context) {
	var payload UserRegistration
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	// Check if the email already exists
	if h.userService.IsEmailExists(payload.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
		return
	}

	// Assuming the default role is 'user'. Modify based on your application logic.
	if err := h.userService.CreateUser(payload.FullName, payload.Email, payload.Password, modeluser.UserRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// / Login handles user login.
func (h *UserHandler) Login(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inputs"})
		return
	}

	// Verify user's password.
	authenticated, err := h.userService.VerifyUserPassword(payload.Email, payload.Password)
	if err != nil || !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Retrieve user details for token generation.
	user, err := h.userService.GetUserByEmail(payload.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user data"})
		return
	}

	// Handle last login time.
	// lastLoginStr := "Never" // Default message for first-time login.
	// if user.LastLogin != nil {
	// 	lastLoginStr = user.LastLogin.Format(time.RFC3339)
	// }

	// Update last login time.
	now := time.Now()
	err = h.userService.UpdateLastLogin(user.ID, &now)
	if err != nil {
		log.Println("Failed to update last login time:", err)
	}

	// JWT token generation.
	jwtWrapper := userauth.JwtWrapper{
		SecretKey:              h.JWTKey,
		Issuer:                 "AuthService",
		AccessTokenExpiration:  userauth.DefaultAccessTokenDuration,
		RefreshTokenExpiration: userauth.DefaultRefreshTokenDuration,
	}
	signedToken, err := jwtWrapper.GenerateToken(user.ID.String(), user.FullName)
	if err != nil {
		log.Println("Error signing token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error signing token"})
		return
	}

	// Generate refresh token.
	signedRefreshToken, err := jwtWrapper.RefreshToken(user.ID.String(), user.FullName)
	if err != nil {
		log.Println("Error signing refresh token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error signing refresh token"})
		return
	}

	// Set refresh token in an HTTP-only cookie.
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refreshToken",
		Value:    signedRefreshToken,
		HttpOnly: true,
		Path:     "/",
		Secure:   false, // Set to true if using HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(jwtWrapper.RefreshTokenExpiration.Seconds()),
	})

	// log.Println("Refresh token:", signedRefreshToken)
	// log.Println("userId:", user.ID.String())

	// Prepare and send the response.
	response := gin.H{
		"id":          user.ID,
		"name":        user.FullName,
		"expiresIn":   time.Now().Add(time.Second * time.Duration(jwtWrapper.AccessTokenExpiration.Seconds())).Unix(),
		"accessToken": signedToken,
		"refreshToken": signedRefreshToken,
		//"lastLogin":   lastLoginStr,
	}
	c.JSON(http.StatusOK, response)
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refreshToken"`
}

// RenewAccessToken handles the renewal of the access token using the refresh token.
func (h *UserHandler) RenewAccessToken(c *gin.Context) {
	log.Println("RenewAccessToken")
	var input RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	refreshToken := input.RefreshToken
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token required"})
		return
	}
	jwtWrapper := userauth.JwtWrapper{
		SecretKey:             h.JWTKey,
		Issuer:                "AuthService",
		AccessTokenExpiration: userauth.DefaultAccessTokenDuration,
	}
	claims, err := jwtWrapper.ValidateToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if the user's account is active
	isSoftDeleted, err := h.userService.IsSoftDeleted(userID)
	if err != nil || isSoftDeleted {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is not active. Please restore to continue."})
		return
	}

	newAccessToken, err := jwtWrapper.GenerateToken(claims.UserID, claims.FullName)
	if err != nil {
		log.Println("Error signing new token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error signing new token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"accessToken": newAccessToken,
		"expiresIn":   time.Now().Add(time.Second * time.Duration(jwtWrapper.AccessTokenExpiration.Seconds())).Unix(),
	})
}

// UpdateProfile handles updating user information.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var payload UserUpdatePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, err := getUserIDFromContext(c)
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

	// Check if the new email already exists for another user
	if payload.Email != user.Email && h.userService.IsEmailExistsForOtherUser(payload.Email, userID) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use."})
		return
	}

	// Update the user data with the payload
	if err := h.userService.UpdateUser(userID, payload.FullName, payload.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteProfile handles the soft deletion of a user profile
func (h *UserHandler) DeleteProfile(c *gin.Context) {
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

// Profile retrieves the user's profile information.
func (h *UserHandler) Profile(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
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

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, errors.New("user ID not provided")
	}

	return uuid.Parse(userIDStr.(string))
}
