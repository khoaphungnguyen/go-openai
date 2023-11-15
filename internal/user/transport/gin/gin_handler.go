package ginuser

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khoaphungnguyen/go-openai/internal/user/auth"
	modeluser "github.com/khoaphungnguyen/go-openai/internal/user/model"
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

// / Signup handles new user registration
func (h *UserHanlder) Signup(c *gin.Context) {
	var user modeluser.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.userService.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

// Login is a function that handles user login
func (h *UserHanlder) Login(c *gin.Context) {
	var payload LoginPayload
	//var user usermodel.User
	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid Inputs",
		})
		c.Abort()
		return
	}
	user, err := h.userService.GetUserByEmail(payload.Email)
	if err != nil {
		c.JSON(401, gin.H{
			"Error": "Invalid Username",
		})
		c.Abort()
		return
	}
	err = user.CheckPassword(payload.Password)
	if err != nil {
		log.Println(err)
		c.JSON(401, gin.H{
			"Error": "Invalid User Credentials",
		})
		c.Abort()
		return
	}
	jwtWrapper := auth.JwtWrapper{
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
func (h *UserHanlder) UpdateProfile(c *gin.Context) {
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

func (h *UserHanlder) DeleteProfile(c *gin.Context) {
	emailInterface, exists := c.Get("email") // Retrieve the email from the context
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email not provided"})
		return
	}

	email, ok := emailInterface.(string) // Type assertion
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email is not a valid string"})
		return
	}

	if err := h.userService.DeleteUser(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}


// Profile retrieves the user's profile information
func (h *UserHanlder) Profile(c *gin.Context) {
	userID, _ := c.Get("userID") // Retrieve the user ID from the context

	user, err := h.userService.GetUserByUUID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Renew token from the refresh token
func (h *UserHanlder) RenewAccessToken(c *gin.Context) {
	token, err := c.Cookie("refreshToken")
	if err != nil {
		c.JSON(400, gin.H{
			"Error": "Invalid Inputs",
		})
		c.Abort()
		return
	}
	jwtWrapper := auth.JwtWrapper{
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
