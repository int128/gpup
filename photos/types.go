package photos

import (
	"net/http"

	"github.com/int128/gpup/photos/internal"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/photoslibrary/v1"
)

// Endpoint is an URL of Google Photos Library API.
var Endpoint = google.Endpoint

// Scopes is a set of OAuth scopes.
var Scopes = []string{photoslibrary.PhotoslibraryScope}

// Photos provides service for manage albums and uploading media items.
type Photos struct {
	service internal.Photos
}

// New creates a Photos.
func New(client *http.Client) (*Photos, error) {
	service, err := internal.New(client)
	if err != nil {
		return nil, err
	}
	return &Photos{service}, nil
}
