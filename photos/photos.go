package photos

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"google.golang.org/api/photoslibrary/v1"
)

const batchCreateSize = 50

// Photos provides service for manage albums and uploading files.
type Photos struct {
	client  *http.Client
	service *photoslibrary.Service
	log     *log.Logger
}

// New creates a Photos.
func New(client *http.Client) (*Photos, error) {
	service, err := photoslibrary.New(client)
	if err != nil {
		return nil, err
	}
	return &Photos{
		client:  client,
		service: service,
		log:     log.New(os.Stdout, "", log.LstdFlags),
	}, nil
}

// AddToLibrary adds the files to the library.
// This method tries uploading all files and ignores any error.
// If no file could be uploaded, this method returns an error.
func (p *Photos) AddToLibrary(filepaths []string) error {
	mediaItems := p.UploadFiles(filepaths)
	if len(mediaItems) == 0 {
		return fmt.Errorf("Could not upload any file")
	}
	for _, chunk := range splitMediaItems(mediaItems, batchCreateSize) {
		p.log.Printf("Adding %d file(s) to the library", len(chunk))
		if err := p.Append(nil, chunk); err != nil {
			return err
		}
	}
	return nil
}

// CreateAlbum creates an album with the files.
// This method tries uploading all files and ignores any error.
// If no file could be uploaded, this method returns an error.
func (p *Photos) CreateAlbum(title string, filepaths []string) (*photoslibrary.Album, error) {
	mediaItems := p.UploadFiles(filepaths)
	if len(mediaItems) == 0 {
		return nil, fmt.Errorf("Could not upload any file")
	}

	p.log.Printf("Creating album %s", title)
	album, err := p.service.Albums.Create(&photoslibrary.CreateAlbumRequest{
		Album: &photoslibrary.Album{
			Title: title,
		},
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("Error while creating an album: %s", err)
	}

	for _, chunk := range splitMediaItems(mediaItems, batchCreateSize) {
		p.log.Printf("Adding %d file(s) into the album %s", len(chunk), album.Title)
		if err := p.Append(album, chunk); err != nil {
			return nil, err
		}
	}
	return album, nil
}
