package oauth

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

// NewConfigForCLI returns a config for CLI interaction.
func NewConfigForCLI(clientID string, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{photoslibrary.PhotoslibraryScope},
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}
}

// GetTokenViaCLI returns a token by browser interaction.
func GetTokenViaCLI(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
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
		return nil, fmt.Errorf("Could not exchange oauth code: %s", err)
	}
	return token, nil
}
