# Use a base image with the Go runtime
FROM golang:latest as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to download the dependencies
COPY go.mod go.sum ./

# Download the Go modules
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app, assuming your entrypoint is cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-openai ./cmd/main.go

# Use a small base image to start a new build stage
FROM alpine:latest

# Install ca-certificates in case your application makes outgoing HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the compiled binary from the builder stage
COPY --from=builder /go-openai /go-openai

# Expose the port the app runs on
EXPOSE 8000

# Run the binary
CMD ["/go-openai"]
