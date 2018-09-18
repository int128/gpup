// Package internal is a thin wrapper of Google Photos Library API,
// providing error handling and retrying.
package internal

import (
	"log"
	"net/http"
	"os"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

// Photos provides Google Photos Library service.
type Photos interface {
	uploadService
	albumsService
	mediaItemsService
}

type defaultPhotos struct {
	client  *http.Client
	service *photoslibrary.Service
	log     *log.Logger
}

// New returns a new Photos.
func New(client *http.Client) (Photos, error) {
	service, err := photoslibrary.New(client)
	if err != nil {
		return nil, err
	}
	return &defaultPhotos{
		client:  client,
		service: service,
		log:     log.New(os.Stderr, "", log.LstdFlags),
	}, nil
}
