package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	chatbusiness "github.com/khoaphungnguyen/go-openai/internal/chat/business"
	chatstorage "github.com/khoaphungnguyen/go-openai/internal/chat/storage"
	chatgin "github.com/khoaphungnguyen/go-openai/internal/chat/transport/gin"
	middleware "github.com/khoaphungnguyen/go-openai/internal/middlewares"
	userbusiness "github.com/khoaphungnguyen/go-openai/internal/user/business"
	userstorage "github.com/khoaphungnguyen/go-openai/internal/user/storage"
	usergin "github.com/khoaphungnguyen/go-openai/internal/user/transport/gin"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
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

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// User and Chat service setup
	userService := userbusiness.NewUserService(userstorage.NewUserStore(db))
	userHandler := usergin.NewUserHandler(userService, jwtKey)

	chatService := chatbusiness.NewChatService(chatstorage.NewChatStore(db))
	chatHandler := chatgin.NewChatHandler(chatService)

	router := gin.Default()
	setupRoutes(router, userHandler, chatHandler, jwtKey)

	if err := router.Run(":8000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func setupRoutes(router *gin.Engine, userHandler *usergin.UserHandler, chatHandler *chatgin.ChatHandler, jwtKey string) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", userHandler.Login)
		auth.POST("/signup", userHandler.Signup)
		auth.POST("/token/renew", userHandler.RenewAccessToken)
	}

	protected := router.Group("/protected").Use(middleware.AuthMiddleware(jwtKey))
	{
		protected.GET("/profile", userHandler.Profile)
		protected.PUT("/profile", userHandler.UpdateProfile)
		protected.PUT("/profile/restore", userHandler.RestoreProfile)
		protected.DELETE("/profile", userHandler.DeleteProfile)

		// Chat routes under protected group
		protected.POST("/thread", chatHandler.CreateThread)
		protected.GET("/thread/:id", chatHandler.GetThread)
		protected.GET("/threads", chatHandler.GetAllThreads)
		protected.DELETE("/thread/:id", chatHandler.DeleteThread)
		protected.POST("/message", chatHandler.CreateMessage)
		protected.GET("/threads/:threadID", chatHandler.GetMessagesByThreadID)
	}
}
