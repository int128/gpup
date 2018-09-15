package photos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/photoslibrary/v1"
)

const batchCreateSize = 20

// Endpoint is an URL of Google Photos Library API.
var Endpoint = google.Endpoint

// Scopes is a set of OAuth scopes.
var Scopes = []string{photoslibrary.PhotoslibraryScope}

// Photos provides service for manage albums and uploading media items.
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
		log:     log.New(os.Stderr, "", log.LstdFlags),
	}, nil
}

// AddToLibrary adds the items to the library.
// This method tries uploading all items and ignores any error.
// If no item could be uploaded, this method returns an error.
func (p *Photos) AddToLibrary(ctx context.Context, mediaItems []MediaItem) error {
	newMediaItems := p.UploadMediaItems(ctx, mediaItems)
	if len(newMediaItems) == 0 {
		return fmt.Errorf("Could not upload any item")
	}
	for _, chunk := range splitMediaItems(newMediaItems, batchCreateSize) {
		p.log.Printf("Adding %d item(s) to the library", len(chunk))
		batch := &photoslibrary.BatchCreateMediaItemsRequest{NewMediaItems: chunk}
		if err := p.BatchCreate(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

// AddToAlbum adds the items to the album.
// This method tries uploading all items and ignores any error.
// If no item could be uploaded, this method returns an error.
func (p *Photos) AddToAlbum(ctx context.Context, title string, mediaItems []MediaItem) error {
	p.log.Printf("Finding album %s", title)
	album, err := p.FindAlbumByTitle(ctx, title)
	if err != nil {
		return fmt.Errorf("Could not list albums: %s", err)
	}
	if album == nil {
		p.log.Printf("Creating album %s", title)
		created, err := p.service.Albums.Create(&photoslibrary.CreateAlbumRequest{
			Album: &photoslibrary.Album{
				Title: title,
			},
		}).Do()
		if err != nil {
			return fmt.Errorf("Error while creating an album: %s", err)
		}
		album = created
	}

	newMediaItems := p.UploadMediaItems(ctx, mediaItems)
	if len(newMediaItems) == 0 {
		return fmt.Errorf("Could not upload any item")
	}
	for _, chunk := range splitMediaItems(newMediaItems, batchCreateSize) {
		p.log.Printf("Adding %d item(s) to the album", len(chunk))
		batch := &photoslibrary.BatchCreateMediaItemsRequest{
			NewMediaItems: chunk,
			AlbumId:       album.Id,
			AlbumPosition: &photoslibrary.AlbumPosition{Position: "LAST_IN_ALBUM"},
		}
		if err := p.BatchCreate(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

// CreateAlbum creates an album with the media items.
// This method tries uploading all items and ignores any error.
// If no item could be uploaded, this method returns an error.
func (p *Photos) CreateAlbum(ctx context.Context, title string, mediaItems []MediaItem) (*photoslibrary.Album, error) {
	newMediaItems := p.UploadMediaItems(ctx, mediaItems)
	if len(newMediaItems) == 0 {
		return nil, fmt.Errorf("Could not upload any item")
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

	for _, chunk := range splitMediaItems(newMediaItems, batchCreateSize) {
		p.log.Printf("Adding %d item(s) into the album %s", len(chunk), album.Title)
		batch := &photoslibrary.BatchCreateMediaItemsRequest{
			NewMediaItems: chunk,
			AlbumId:       album.Id,
			AlbumPosition: &photoslibrary.AlbumPosition{Position: "LAST_IN_ALBUM"},
		}
		if err := p.BatchCreate(ctx, batch); err != nil {
			return nil, err
		}
	}
	return album, nil
}
