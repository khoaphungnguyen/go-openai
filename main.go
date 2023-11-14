package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// Get the OpenAI API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("The environment variable OPENAI_API_KEY is required")
		return
	}

	// Initialize OpenAI client
	client := openai.NewClient(apiKey)

	// Initialize the Chi router
	r := chi.NewRouter()

	// Define the endpoint for chat
	r.Get("/chat", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Transfer-Encoding", "chunked")

		// Set up a channel to receive messages from OpenAI
		messageChan := make(chan string)

		// Start a go routine to stream from OpenAI
		go func() {
			defer close(messageChan)
			for {
				select {
				case <-r.Context().Done():
					return
				default:
					responseContent, err := chatHandler(client, "Your prompt here")
					if err != nil {
						fmt.Printf("Error calling OpenAI: %v\n", err)
						return
					}
					messageChan <- responseContent
				}
			}
		}()

		// Stream messages to the client
		for {
			select {
			case <-r.Context().Done():
				return
			case message := <-messageChan:
				fmt.Fprintf(w, "data: %s\n\n", message)
				flusher.Flush()
			}
		}
	})

	// Start the HTTP server
	fmt.Println("The server is started on port 8000")
	http.ListenAndServe(":8000", r)
}

// callOpenAI interacts with the OpenAI API and returns the response content
func chatHandler(client *openai.Client, userMessage string) (string, error) {
	// Set up the chat completion request
	req := openai.ChatCompletionRequest{
		Model:     "gpt-3.5-turbo",
		MaxTokens: 20,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: userMessage,
			},
		},
	}

	// Create the chat completion stream
	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	// Read the response from the stream
	saveContent := ""
	for {
		response, err := stream.Recv()
		if err != nil {
			break
		}
		saveContent += response.Choices[0].Delta.Content
	}

	// Return the concatenated responses
	return saveContent, nil
}
