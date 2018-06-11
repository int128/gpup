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

const httpPort = 8000

// NewConfigForBrowser returns a config for browser interaction.
func NewConfigForBrowser(clientID string, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{photoslibrary.PhotoslibraryScope},
		RedirectURL:  fmt.Sprintf("http://localhost:%d/", httpPort),
	}
}

// GetTokenViaBrowser returns a token by browser interaction.
func GetTokenViaBrowser(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	state, err := generateOAuthState()
	if err != nil {
		return nil, err
	}
	log.Printf("Open http://localhost:%d for authorization", httpPort)
	code, err := getCodeViaBrowser(ctx, config, state)
	if err != nil {
		return nil, err
	}
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("Could not exchange oauth code: %s", err)
	}
	return token, nil
}

func getCodeViaBrowser(ctx context.Context, config *oauth2.Config, state string) (string, error) {
	codeCh := make(chan string)
	errCh := make(chan error)
	server := http.Server{
		Addr: fmt.Sprintf(":%d", httpPort),
		Handler: &AuthCodeGrantHandler{
			AuthCodeURL: config.AuthCodeURL(state),
			Callback: func(code string, actualState string, err error) {
				switch {
				case err != nil:
					errCh <- err
				case actualState != state:
					errCh <- fmt.Errorf("OAuth state did not match, should be %s but %s", state, actualState)
				default:
					codeCh <- code
				}
			},
		},
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	select {
	case err := <-errCh:
		server.Shutdown(ctx)
		return "", err
	case code := <-codeCh:
		server.Shutdown(ctx)
		return code, nil
	}
}

// AuthCodeGrantHandler handles requests for OIDC auth code grant
type AuthCodeGrantHandler struct {
	AuthCodeURL string
	Callback    func(code string, state string, err error)
}

func (s *AuthCodeGrantHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.RequestURI)
	switch r.URL.Path {
	case "/":
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		errorCode := r.URL.Query().Get("error")
		errorDescription := r.URL.Query().Get("error_description")
		switch {
		case code != "":
			s.Callback(code, state, nil)
			fmt.Fprintf(w, "Back to command line.")
		case errorCode != "":
			s.Callback("", "", fmt.Errorf("OAuth Error: %s %s", errorCode, errorDescription))
			fmt.Fprintf(w, "Back to command line.")
		default:
			http.Redirect(w, r, s.AuthCodeURL, 302)
		}
	default:
		http.Error(w, "Not Found", 404)
	}
}
