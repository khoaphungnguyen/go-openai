package openaitransport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khoaphungnguyen/go-openai/internal/common"
	openaimodel "github.com/khoaphungnguyen/go-openai/internal/openai/model"
	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

type LocalMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LocalChat struct {
	Model    string         `json:"model"`
	Messages []LocalMessage `json:"messages"`
	Stream   bool           `json:"stream"`
	Format   string         `json:"format"`
}

type LocalOptions struct {
	NumPredict int `json:"num_predict"`
}
type LocalRequest struct {
	Model   string       `json:"model"`
	Prompt  string       `json:"prompt"`
	System  string       `json:"system"`
	Stream  bool         `json:"stream"`
	Options LocalOptions `json:"options"`
}

type Response struct {
	Model     string       `json:"model"`
	CreatedAt string       `json:"created_at"`
	Message   LocalMessage `json:"message"`
	Done      bool         `json:"done"`
}

type LocalResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

type Options struct {
	Temperature float64
	NumPredict  int
	NumCtx      int
}

// CreateTransaction handles the creation of a new OpenAI transaction (HTTP Handler).
func (h *OpenAIHandler) CreateTransaction(c *gin.Context) {
	userID, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	var inputData openaimodel.OpenAITransactionInput
	if err := c.ShouldBindJSON(&inputData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	_, err = h.createTransaction(userID, inputData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction created"})
}

// Core logic for creating a transaction
func (h *OpenAIHandler) createTransaction(userID uuid.UUID, inputData openaimodel.OpenAITransactionInput) (uuid.UUID, error) {
	threadID, err := uuid.Parse(inputData.ThreadID)
	if err != nil {
		return uuid.Nil, err
	}
	return h.openAIService.CreateTransaction(userID, threadID, inputData.Message, inputData.Model, inputData.Role)
}

// GetTransactionsByUserID handles fetching transactions for a specific user.
func (h *OpenAIHandler) GetTransactionsByUserID(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	transactions, err := h.openAIService.GetTransactionsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// UpdateTransaction handles the updating of an existing OpenAI transaction.
func (h *OpenAIHandler) UpdateTransaction(c *gin.Context) {
	var transaction openaimodel.OpenAITransaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.openAIService.UpdateTransaction(&transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction updated"})
}

// DeleteTransaction handles the deletion of an OpenAI transaction.
func (h *OpenAIHandler) DeleteTransaction(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("transactionID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	if err := h.openAIService.DeleteTransaction(transactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted"})
}

// GetTransactionByID handles fetching a specific transaction by its ID.
func (h *OpenAIHandler) GetTransactionByID(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("transactionID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	transaction, err := h.openAIService.GetTransactionByID(transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// FetchSuggestion handles the request to fetch suggestions from OpenAI.
func (h *OpenAIHandler) FetchSuggestion(c *gin.Context) {
	// Extract user ID from context, if required
	_, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Retrieve the OpenAI client from the context
	openaiClient, exists := c.MustGet("openaiClient").(*openai.Client)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI client not available"})
		return
	}
	type RequestData struct {
		Model string `json:"model"`
	}

	// Bind the input data (assumed to be the model name)
	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	if !strings.HasPrefix(requestData.Model, "gpt") {
		// Construct the prompt
		prompt := `Provide four engaging recommendations (max 10 words each) as JSON :: [{ "title": "", "content": "" }, ...]`
		req := LocalRequest{
			Model:  requestData.Model,
			Prompt: prompt,
			Stream: false,
		}
		reqJson, _ := json.Marshal(req)
		resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(reqJson))
		if err != nil {
			// handle error
			log.Println(err)
		}
		defer resp.Body.Close()
		// Create the request payload and send the request
		message, err := io.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Println(err)
		}

		var respData LocalResponse
		err = json.Unmarshal(message, &respData)
		if err != nil {
			// handle error
			log.Println(err)
		}

		log.Println("Response from local server: ", respData.Response)

		// Check if the response has content and return the content
		if len(respData.Response) > 0 {
			c.JSON(http.StatusOK, respData.Response)
			return
		}
	} else {
		// Construct the prompt
		prompt := `Provide only four engaging recommendations (max 10 words each) as JSON 
		: [{ "title": "", "content": "" }, ...]`
		// Create the request payload and send the request to OpenAI
		resp, err := openaiClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:            requestData.Model,
				Messages:         []openai.ChatCompletionMessage{{Role: "user", Content: prompt}},
				Temperature:      1,
				TopP:             1,
				FrequencyPenalty: 0,
				MaxTokens:        200,
				N:                1,
				Stream:           false,
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
			return
		}

		// Check if the response has content and return the content
		if len(resp.Choices) > 0 && len(resp.Choices[0].Message.Content) > 0 {
			c.JSON(http.StatusOK, resp.Choices[0].Message.Content)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No content in response"})
	}
}

// FetchSuggestion handles the request to fetch suggestions from OpenAI.
func (h *OpenAIHandler) GenerateHint(c *gin.Context) {
	// Extract user ID from context, if required
	_, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Retrieve the OpenAI client from the context
	openaiClient, exists := c.MustGet("openaiClient").(*openai.Client)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI client not available"})
		return
	}
	type RequestData struct {
		Model  string `json:"model"`
		Input  string `json:"input"`
		System string `json:"system"`
	}

	// Bind the input data (assumed to be the model name)
	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	if !strings.HasPrefix(requestData.Model, "gpt") {
		// Construct the prompt
		req := LocalRequest{
			Model:  requestData.Model,
			Prompt: requestData.Input,
			System: requestData.System,
			Stream: false,
			Options: LocalOptions{
				NumPredict: 200,
			},
		}
		reqJson, _ := json.Marshal(req)
		resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(reqJson))
		if err != nil {
			// handle error
			log.Println(err)
		}
		defer resp.Body.Close()
		// Create the request payload and send the request
		message, err := io.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Println(err)
		}

		var respData LocalResponse
		err = json.Unmarshal(message, &respData)
		if err != nil {
			// handle error
			log.Println(err)
		}
		// Check if the response has content and return the content
		if len(respData.Response) > 0 {
			c.JSON(http.StatusOK, respData.Response)
			return
		}
	} else {
		// Create the request payload and send the request to OpenAI
		resp, err := openaiClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: requestData.Model,
				Messages: []openai.ChatCompletionMessage{{
					Role: "system", Content: requestData.System,
				}, {Role: "user", Content: requestData.Input}},
				Temperature:      1,
				TopP:             1,
				FrequencyPenalty: 0,
				MaxTokens:        500,
				N:                1,
				Stream:           false,
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
			return
		}

		// Check if the response has content and return the content
		if len(resp.Choices) > 0 && len(resp.Choices[0].Message.Content) > 0 {
			c.JSON(http.StatusOK, resp.Choices[0].Message.Content)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No content in response"})
	}
}

// FetchSuggestion handles the request to fetch suggestions from OpenAI.
func (h *OpenAIHandler) FetchDrawing(c *gin.Context) {
	// Extract user ID from context, if required
	_, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Retrieve the OpenAI client from the context
	openaiClient, exists := c.MustGet("openaiClient").(*openai.Client)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI client not available"})
		return
	}

	type RequestData struct {
		ImageURL string `json:"imageURL"`
	}

	// Bind the input data (assumed to be the model name)
	var requestData RequestData
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err != nil {
		log.Println("Error from OpenAI: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
		return
	}
	// Construct the prompt
	systemPrompt := `You are an expert Tailwind developer. A user will provide you with a low-fidelity wireframe of an application and you will return a single html file that uses Tailwind to create the website. Use creative license to make the application more fleshed out. If you need to insert an image, use placehold.co to create a placeholder image. Respond only with the html file.`

	// Create the request payload and send the request to OpenAI
	resp, err := openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4-vision-preview",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: systemPrompt,
				},
				{
					Role: "user",
					MultiContent: []openai.ChatMessagePart{
						{
							Type: "image_url",
							ImageURL: &openai.ChatMessageImageURL{
								URL:    requestData.ImageURL,
								Detail: openai.ImageURLDetailAuto,
							},
						},
					},
				},
			},
			MaxTokens: 1000,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
		return
	}
	// Check if the response has content and return the content
	if len(resp.Choices) > 0 && len(resp.Choices[0].Message.Content) > 0 {
		c.JSON(http.StatusOK, resp.Choices[0].Message.Content)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "No content in response"})
}

type MessageInput struct {
	Messages []LocalMessage `json:"messages"`
	Model    string         `json:"model"`
}

// MessageHandler handles the incoming messages.
func (h *OpenAIHandler) MessageHanlder(c *gin.Context) {
	userID, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, "Invalid thread ID")
		return
	}
	// Check some condition to decide whether to stop the generation
	if c.Query("stop") == "true" {
		err := h.StopGeneration(threadID)
		if err != nil {
			log.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{"message": "Generation stopped"})
		return
	}

	openaiClient, exists := c.MustGet("openaiClient").(*openai.Client)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI client not available"})
		return
	}

	// Binding the request data
	var inputData MessageInput
	if err := c.ShouldBindJSON(&inputData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// add all messages to the prompt
	var prompt strings.Builder
	for _, message := range inputData.Messages {
		prompt.WriteString(message.Content)
	}
	//println("Prompt: ", prompt.String())
	encoding := "cl100k_base"
	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return
	}
	// encodd
	token := tke.Encode(prompt.String(), nil, nil)

	println("Tokens from the prompt	 ", len(token))

	var message string
	if len(inputData.Messages) > 0 {
		// Get the content of the last message
		message = inputData.Messages[len(inputData.Messages)-1].Content
	}

	// Log the user's message
	transactionID, err := h.createTransaction(userID, openaimodel.OpenAITransactionInput{
		ThreadID: threadID.String(),
		Message:  message,
		Model:    inputData.Model,
		Role:     "user",
	})
	if err != nil {
		log.Printf("Error saving user transaction: %v", err)
		return
	}

	transaction, err := h.openAIService.GetTransactionByID(transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction"})
		return
	}
	log.Println("Tokens from the prompt	 ", transaction.MessageLength)

	// Create a new context with a cancel function
	ctx, cancel := context.WithCancel(context.Background())
	//Set the cancel function for the thread ID
	h.Mutex.Lock()
	h.CancelFuncsLLM[threadID] = cancel

	h.Mutex.Unlock()

	if !strings.HasPrefix(inputData.Model, "gpt") {
		chat := LocalChat{
			Model:    inputData.Model,
			Messages: inputData.Messages,
			Stream:   true,
		}
		chatJson, _ := json.Marshal(chat)
		stream, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(chatJson))
		if err != nil {
			// handle error
			log.Println(err)
		}
		defer stream.Body.Close()
		// Stream the response from OpenAI and send parts to the client via SSE
		h.localStreamResponse(c, ctx, threadID, userID, inputData.Model, stream)
	} else {
		// Convert []LocalMessage to []openai.ChatCompletionMessage
		var messages []openai.ChatCompletionMessage
		for _, localMessage := range inputData.Messages {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    localMessage.Role,
				Content: localMessage.Content,
			})
		}

		// Set up the chat completion request
		req := openai.ChatCompletionRequest{
			Model:       inputData.Model,
			Stream:      true,
			MaxTokens:   1000,
			Messages:    messages,
			Temperature: 0,
			TopP:        0.9,
		}
		// Create chat completion stream
		stream, err := openaiClient.CreateChatCompletionStream(h.ctx, req) // Use the handler's context
		if err != nil {
			log.Printf("CreateChatCompletionStream error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stream"})
			return
		}
		defer stream.Close()
		// Stream the response from OpenAI and send parts to the client via SSE
		h.openAIStreamResponse(c, ctx, threadID, userID, inputData.Model, stream)
	}
}

func (h *OpenAIHandler) localStreamResponse(c *gin.Context, ctx context.Context, threadID uuid.UUID, userID uuid.UUID, model string, stream *http.Response) {

	var responseBuilder strings.Builder

	// Defer the saving logic so it always runs, even if the function returns early
	defer func() {
		transactionID, err := h.createTransaction(userID, openaimodel.OpenAITransactionInput{
			ThreadID: threadID.String(),
			Message:  responseBuilder.String(),
			Model:    model,
			Role:     "assistant",
		})
		if err != nil {
			log.Printf("Error saving assistant transaction: %v", err)
		}
		transaction, err := h.openAIService.GetTransactionByID(transactionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction"})
			return
		}
		log.Println("Tokens from Response: ", transaction.MessageLength)
		c.JSON(http.StatusOK, gin.H{"message": "Message received and processed"})
	}()

	// Create a new JSON decoder for the response body
	decoder := json.NewDecoder(stream.Body)

	// Add a label for the for loop
loop:
	for {
		select {
		case <-ctx.Done():
			// If the context has been cancelled, stop reading from the stream
			ch, exists := h.ThreadSSEChannels[threadID]
			if !exists {
				ch = make(chan string, 100)
				h.ThreadSSEChannels[threadID] = ch
			}
			ch <- ""
			return
		default:
			// If the context has not been cancelled, read the next line from the stream
			var response Response
			if err := decoder.Decode(&response); err != nil {
				if err == io.EOF {
					// End of the stream, break the loop
					break loop
				} else {
					// handle error
					log.Println(err)
					break loop
				}
			}

			responseContent := response.Message.Content
			responseBuilder.WriteString(responseContent)

			ch, exists := h.ThreadSSEChannels[threadID]
			if !exists {
				ch = make(chan string, 100)
				h.ThreadSSEChannels[threadID] = ch
			}

			select {
			case ch <- responseContent:
				// Successfully sent to channel
			default:
				log.Printf("Channel buffer full or closed. Dropping message for thread ID %s.", threadID)
			}
		}
	}
}

func (h *OpenAIHandler) openAIStreamResponse(c *gin.Context, ctx context.Context, threadID uuid.UUID, userID uuid.UUID, model string, stream *openai.ChatCompletionStream) {
	var responseBuilder strings.Builder

	// Start a goroutine to do the processing in the background

loop:
	for {
		select {
		case <-ctx.Done():
			// If the context has been cancelled, stop reading from the stream
			ch, exists := h.ThreadSSEChannels[threadID]
			if !exists {
				ch = make(chan string, 100)
				h.ThreadSSEChannels[threadID] = ch
			}
			ch <- ""
			return
		default:
			// If the context has not been cancelled, read the next line from the stream
			response, err := stream.Recv()
			if err == io.EOF {
				break loop
			} else if err != nil {
				break loop
			}

			responseContent := response.Choices[0].Delta.Content
			responseBuilder.WriteString(responseContent)

			ch, exists := h.ThreadSSEChannels[threadID]
			if !exists {
				ch = make(chan string, 100)
				h.ThreadSSEChannels[threadID] = ch
			}

			select {
			case ch <- responseContent:
				// Successfully sent to channel
			default:
				log.Printf("Channel buffer full or closed. Dropping message for thread ID %s.", threadID)
			}
		}
	}

	// Start another goroutine to run createTransaction after the loop has finished
	defer func() {
		transactionID, err := h.createTransaction(userID, openaimodel.OpenAITransactionInput{
			ThreadID: threadID.String(),
			Message:  responseBuilder.String(),
			Model:    model,
			Role:     "assistant",
		})
		if err != nil {
			log.Printf("Error saving assistant transaction: %v", err)
		}

		transaction, err := h.openAIService.GetTransactionByID(transactionID)
		if err != nil {
			log.Printf("Failed to get transaction: %v", err)
			return
		}
		log.Println("Tokens from Response: ", transaction.MessageLength)
	}()

	// Send a response to the client immediately
	c.JSON(http.StatusOK, gin.H{"message": "Message received and processing started. Please check back later for the result."})
}

func (h *OpenAIHandler) StopGeneration(threadID uuid.UUID) error {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	if cancel, ok := h.CancelFuncsLLM[threadID]; ok {
		cancel()
		delete(h.CancelFuncs, threadID)
	} else {
		return fmt.Errorf("no cancel function for thread ID %v", threadID)
	}

	return nil
}

func (h *OpenAIHandler) SSEHandler(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("threadID"))
	if err != nil {
		common.RespondWithError(c, http.StatusBadRequest, "Invalid thread ID")
		return
	}

	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// Ensure the channel exists and return the channel
	ch := h.ensureChannelExists(threadID, c.Writer)

	// Just ensuring the channel and goroutine are set up correctly
	if ch == nil {
		common.RespondWithError(c, http.StatusInternalServerError, "Failed to set up SSE stream")
		return
	}

	// Keep the connection open until the client closes it
	<-c.Request.Context().Done()
	log.Println("Client closed connection")

	// Cancel the receiving goroutine when the client closes the connection
	if cancel, exists := h.CancelFuncs[threadID]; exists {
		cancel()
		delete(h.CancelFuncs, threadID)
	}
}

func (h *OpenAIHandler) ensureChannelExists(threadID uuid.UUID, w http.ResponseWriter) chan string {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	ch, exists := h.ThreadSSEChannels[threadID]
	if !exists {
		ch = make(chan string, 100)
		h.ThreadSSEChannels[threadID] = ch
	}

	// Cancel the old goroutine if it exists
	if cancel, exists := h.CancelFuncs[threadID]; exists {
		cancel()
		delete(h.CancelFuncs, threadID)
	}

	// Create a new context with a cancel function
	ctx, cancel := context.WithCancel(h.ctx)

	// Store the cancel function
	h.CancelFuncs[threadID] = cancel

	h.startReceiving(ctx, threadID, ch, w) // Always start a new goroutine

	return ch
}

func (h *OpenAIHandler) startReceiving(ctx context.Context, threadID uuid.UUID, ch chan string, w http.ResponseWriter) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Println("Streaming not supported")
		return
	}

	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return // Channel was closed, exit the goroutine.
				}

				messageID := uuid.New().String()
				createdAt := time.Now().Format(time.RFC3339)
				jsonResponse := fmt.Sprintf(`{"id": %q, "content": %q, "role": "assistant", "createdAt": %q}`, messageID, msg, createdAt)
				fmt.Fprintf(w, "data: %s\n\n", jsonResponse)
				// Check go rountine ID

				//log.Println("Sent response to client: ", msg)
				flusher.Flush()

			case <-ctx.Done():
				return // Context was cancelled, exit the goroutine.
			}
		}
	}()
}

// func getGoroutineID() uint64 {
// 	var buf [64]byte
// 	n := runtime.Stack(buf[:], false)
// 	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
// 	id, err := strconv.ParseUint(idField, 10, 64)
// 	if err != nil {
// 		log.Printf("Failed to parse goroutine id: %v", err)
// 		return 0
// 	}
// 	return id
// }
