// main initializes and runs the chat application server.
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	messagebusiness "github.com/khoaphungnguyen/go-openai/internal/message/business"
	messagestorage "github.com/khoaphungnguyen/go-openai/internal/message/storage"
	messagetransport "github.com/khoaphungnguyen/go-openai/internal/message/transport"
	middleware "github.com/khoaphungnguyen/go-openai/internal/middlewares"
	notebusiness "github.com/khoaphungnguyen/go-openai/internal/note/business"
	notestorage "github.com/khoaphungnguyen/go-openai/internal/note/storage"
	notetransport "github.com/khoaphungnguyen/go-openai/internal/note/transport"
	openaibusiness "github.com/khoaphungnguyen/go-openai/internal/openai/business"
	openaistorage "github.com/khoaphungnguyen/go-openai/internal/openai/storage"
	openaitransport "github.com/khoaphungnguyen/go-openai/internal/openai/transport"
	userbusiness "github.com/khoaphungnguyen/go-openai/internal/user/business"
	userstorage "github.com/khoaphungnguyen/go-openai/internal/user/storage"
	usertransport "github.com/khoaphungnguyen/go-openai/internal/user/transport"
)

func main() {
	// Application setup and route configuration...
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

	// Initialize OpenAI client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}
	openaiClient := openai.NewClient(apiKey)

	// User and Chat service setup
	userService := userbusiness.NewUserService(userstorage.NewUserStore(db))
	userHandler := usertransport.NewUserHandler(userService, jwtKey)

	messageService := messagebusiness.NewMessageService(messagestorage.NewMessageStore(db))
	messageHandler := messagetransport.NewMessageHandler(messageService)

	noteService := notebusiness.NewNoteService(notestorage.NewNoteStore(db))
	noteHandler := notetransport.NewNoteHandler(noteService)

	openaiService := openaibusiness.NewOpenAIService(openaistorage.NewOpenAIStore(db), messageService)
	chatHandler := openaitransport.NewOpenAIHandler(openaiService)

	router := gin.Default()
	router.Use(middleware.CORSMiddleware([]string{"http://localhost:3000"}))
	setupRoutes(router, userHandler, messageHandler, noteHandler, chatHandler, jwtKey, openaiClient)

	if err := router.Run(":8000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

// setupRoutes defines the HTTP routes for the application.
func setupRoutes(router *gin.Engine, userHandler *usertransport.UserHandler, messageHandler *messagetransport.MessageHandler, noteHandler *notetransport.NoteHandler,
	openAIHandler *openaitransport.OpenAIHandler, jwtKey string, openaiClient *openai.Client) {

	auth := router.Group("/auth")
	{
		auth.POST("/login", userHandler.Login)
		auth.POST("/signup", userHandler.Signup)
		auth.POST("/refresh", userHandler.RenewAccessToken)
	}

	protected := router.Group("/protected").Use(middleware.AuthMiddleware(jwtKey))
	{
		protected.GET("/users", userHandler.GetAllUsers)
		protected.GET("/profile", userHandler.Profile)
		protected.PUT("/profile", userHandler.UpdateProfile)
		protected.PUT("/profile/restore", userHandler.RestoreProfile)
		protected.DELETE("/profile", userHandler.DeleteProfile)

		// ChatMessage routes under protected group
		protected.POST("/thread", messageHandler.CreateThread)
		protected.GET("/thread/:id", messageHandler.GetThreadByID)
		protected.GET("/threads", messageHandler.GetAllThreads)
		protected.DELETE("/thread/:id", messageHandler.DeleteThread)
		protected.POST("/message", messageHandler.CreateMessage)
		protected.GET("/threads/:threadID", messageHandler.GetMessagesByThreadID)

		// Note routes under protected group
		protected.POST("/notes", noteHandler.CreateNote)
		protected.GET("/notes", noteHandler.GetAllNoteByUserID)
		protected.GET("/notes/:noteID", noteHandler.GetNoteByID)
		protected.PUT("/notes/:noteID", noteHandler.UpdateNote)
		protected.DELETE("/notes/:noteID", noteHandler.DeleteNote)

		// Apply OpenAIClientMiddleware to the protected group that requires OpenAI client
		protected.Use(middleware.OpenAIClientMiddleware(openaiClient))
		protected.POST("/suggestions", openAIHandler.FetchSuggestion)
		protected.POST("/drawings", openAIHandler.FetchDrawing)
		protected.POST("/transactions", openAIHandler.CreateTransaction)
		protected.GET("/transactions/user/:userID", openAIHandler.GetTransactionsByUserID)
		protected.PUT("/transactions", openAIHandler.UpdateTransaction)
		protected.DELETE("/transactions/:transactionID", openAIHandler.DeleteTransaction)
		protected.GET("/transactions/:transactionID", openAIHandler.GetTransactionByID)
		protected.GET("/chat/:threadID", openAIHandler.WebSocketHandler)
		protected.GET("/chat/stream/:threadID", openAIHandler.SSEHandler)
		protected.POST("/chat/ask/:threadID", openAIHandler.MessageHanlder)
	}
}
