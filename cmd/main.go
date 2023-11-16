package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	middleware "github.com/khoaphungnguyen/go-openai/internal/middlewares"
	userbusiness "github.com/khoaphungnguyen/go-openai/internal/user/business"
	userstorage "github.com/khoaphungnguyen/go-openai/internal/user/storage"
	usergin "github.com/khoaphungnguyen/go-openai/internal/user/transport/gin"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Fetch JWT key and database configuration from the environment
	jwtKey := os.Getenv("JWT_SECRET_KEY")
	if jwtKey == "" {
		log.Fatal("JWT_SECRET_KEY not set in .env file")
	}

	dbURL := os.Getenv("DATABASE_LOCAL_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_LOCAL_URL not set in .env file")
	}

	// Initialize database connection
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create a new user storage instance
	userStore := userstorage.NewUserStore(db)

	// Create a new user service
	userService := userbusiness.NewUserService(userStore)
	userHandler := usergin.NewUserHandler(userService, jwtKey)

	r := setupRouter(userHandler)
	if err := r.Run(":8000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func setupRouter(userHandler *usergin.UserHandler) *gin.Engine {
	r := gin.Default()

	// Create a new group for the API
	auth := r.Group("/auth")
	{
		auth.POST("/login", userHandler.Login)
		auth.POST("/signup", userHandler.Signup)
		auth.POST("/refresh", userHandler.RenewAccessToken)
	}

	// Create protected route
	protected := r.Group("/protected").Use(middleware.AuthMiddleware(userHandler.JWTKey))
	{
		protected.GET("/profile", userHandler.Profile)
		protected.PUT("/profile", userHandler.UpdateProfile)
		protected.PUT("/profile/restore", userHandler.RestoreProfile)
		protected.DELETE("/profile", userHandler.DeleteProfile)
	}
	return r
}
