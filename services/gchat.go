package services

import (
	"context"
	"fmt"
	"os"
	"sync"

	"google.golang.org/api/chat/v1"
	"google.golang.org/api/option"
)

// NewGChatService creates and initializes a new Google Chat service
func NewGChatService() (*chat.Service, error) {
	ctx := context.Background()

	credentialsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	if credentialsFile == "" {
		panic("GOOGLE_CREDENTIALS_FILE environment variable must be set")
	}

	tokenFile := os.Getenv("GOOGLE_TOKEN_FILE")
	if tokenFile == "" {
		panic("GOOGLE_TOKEN_FILE environment variable must be set")
	}

	client := GoogleHttpClient(tokenFile, credentialsFile)

	// Initialize Google Chat API service with default credentials and required scopes
	srv, err := chat.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create chat service: %v", err)
	}

	return srv, nil
}

var DefaultGChatService = sync.OnceValue[*chat.Service](func() *chat.Service {
	srv, err := NewGChatService()
	if err != nil {
		panic(fmt.Sprintf("failed to create chat service: %v", err))
	}
	return srv
})