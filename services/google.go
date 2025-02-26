package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/youtube/v3"
)

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func ListChatScopes() []string {
	return []string{
		"https://www.googleapis.com/auth/chat.admin.memberships",
		"https://www.googleapis.com/auth/chat.admin.memberships.readonly",
		"https://www.googleapis.com/auth/chat.admin.spaces",
		"https://www.googleapis.com/auth/chat.admin.spaces.readonly",
		"https://www.googleapis.com/auth/chat.memberships",
		"https://www.googleapis.com/auth/chat.memberships.app",
		"https://www.googleapis.com/auth/chat.memberships.readonly",
		"https://www.googleapis.com/auth/chat.messages",
		"https://www.googleapis.com/auth/chat.messages.create",
		"https://www.googleapis.com/auth/chat.messages.reactions",
		"https://www.googleapis.com/auth/chat.messages.reactions.create",
		"https://www.googleapis.com/auth/chat.messages.reactions.readonly",
		"https://www.googleapis.com/auth/chat.messages.readonly",
		"https://www.googleapis.com/auth/chat.spaces",
		"https://www.googleapis.com/auth/chat.spaces.create",
		"https://www.googleapis.com/auth/chat.spaces.readonly",
		"https://www.googleapis.com/auth/chat.users.readstate",
		"https://www.googleapis.com/auth/chat.users.readstate.readonly",
	}
}
func ListGoogleScopes() []string {
	scopes := []string{
		gmail.GmailLabelsScope,
		gmail.GmailModifyScope,
		gmail.MailGoogleComScope,
		gmail.GmailSettingsBasicScope,
		calendar.CalendarScope,
		calendar.CalendarEventsScope,
		youtube.YoutubeScope,
		youtube.YoutubeUploadScope,
		youtube.YoutubepartnerChannelAuditScope,
		youtube.YoutubepartnerScope,
		youtube.YoutubeReadonlyScope,
	}
	scopes = append(scopes, ListChatScopes()...)
	return scopes
}

func GoogleHttpClient(tokenFile string, credentialsFile string) *http.Client {
	
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		panic(fmt.Sprintf("failed to read token file: %v", err))
	}

	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, ListGoogleScopes()...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	return config.Client(ctx, tok)
}