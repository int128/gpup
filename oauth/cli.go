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

// NewClientViaCLI creates an OAuth client via CLI.
func NewClientViaCLI(ctx context.Context, clientID string, clientSecret string) (*http.Client, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{photoslibrary.PhotoslibraryScope},
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}
	cache, err := FindTokenCache(config)
	switch {
	case cache != nil:
		return config.Client(ctx, cache), nil
	case err != nil:
		log.Printf("Could not find token cache: %s", err)
		fallthrough
	default:
		token, err := getTokenViaCLI(ctx, config)
		if err != nil {
			return nil, err
		}
		if err := CreateTokenCache(token, config); err != nil {
			log.Printf("Could not store token cache: %s", err)
		}
		return config.Client(ctx, token), nil
	}
}

func getTokenViaCLI(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
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
