package oauth

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

// NewClientViaCLI creates a new http.Client via CLI.
func NewClientViaCLI(ctx context.Context, clientID string, clientSecret string) (*http.Client, error) {
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
	log.Printf("Open %s for authorization", authCodeURL)
	fmt.Print("Enter code: ")
	var code string
	if _, err := fmt.Scanln(&code); err != nil {
		return nil, err
	}
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return config.Client(ctx, token), nil
}
