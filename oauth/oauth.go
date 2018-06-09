package oauth

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

// NewClient creates a new http.Client with a bearer access token
func NewClient(ctx context.Context, clientID string, clientSecret string) (*http.Client, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{photoslibrary.PhotoslibraryScope},
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}
	state, err := generateOAuthState()
	if err != nil {
		return nil, err
	}
	authCodeURL := config.AuthCodeURL(state)
	log.Printf("Open %s", authCodeURL)
	fmt.Print("Enter code: ")
	var authCode string
	if _, err := fmt.Scanln(&authCode); err != nil {
		return nil, err
	}
	accessToken, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, err
	}
	return config.Client(ctx, accessToken), nil
}

func generateOAuthState() (string, error) {
	var n uint64
	if err := binary.Read(rand.Reader, binary.LittleEndian, &n); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", n), nil
}
