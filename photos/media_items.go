package photos

import (
	"context"
	"fmt"
	"log"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

const batchCreateSize = 20

// AddToLibrary adds the items to the library.
// This method tries uploading all items and ignores any error.
// If no item could be uploaded, this method returns an error.
func (p *Photos) AddToLibrary(ctx context.Context, mediaItems []MediaItem) error {
	newMediaItems := p.UploadMediaItems(ctx, mediaItems)
	if len(newMediaItems) == 0 {
		return fmt.Errorf("Could not upload any item")
	}
	for _, chunk := range splitMediaItems(newMediaItems, batchCreateSize) {
		log.Printf("Adding %d item(s) to the library", len(chunk))
		batch := &photoslibrary.BatchCreateMediaItemsRequest{NewMediaItems: chunk}
		if err := p.service.BatchCreate(ctx, batch); err != nil {
			return fmt.Errorf("Could not add items to the library: %s", err)
		}
	}
	return nil
}

// AddToAlbum adds the items to the album.
// This method tries uploading all items and ignores any error.
// If no item could be uploaded, this method returns an error.
func (p *Photos) AddToAlbum(ctx context.Context, title string, mediaItems []MediaItem) error {
	log.Printf("Finding album %s", title)
	album, err := p.FindAlbumByTitle(ctx, title)
	if err != nil {
		return fmt.Errorf("Could not list albums: %s", err)
	}
	if album == nil {
		log.Printf("Creating album %s", title)
		created, err := p.service.CreateAlbum(ctx, &photoslibrary.CreateAlbumRequest{
			Album: &photoslibrary.Album{Title: title},
		})
		if err != nil {
			return fmt.Errorf("Could not create an album: %s", err)
		}
		album = created
	}

	newMediaItems := p.UploadMediaItems(ctx, mediaItems)
	if len(newMediaItems) == 0 {
		return fmt.Errorf("Could not upload any item")
	}
	for _, chunk := range splitMediaItems(newMediaItems, batchCreateSize) {
		log.Printf("Adding %d item(s) to the album", len(chunk))
		batch := &photoslibrary.BatchCreateMediaItemsRequest{
			NewMediaItems: chunk,
			AlbumId:       album.Id,
			AlbumPosition: &photoslibrary.AlbumPosition{Position: "LAST_IN_ALBUM"},
		}
		if err := p.service.BatchCreate(ctx, batch); err != nil {
			return fmt.Errorf("Could not add items to the album: %s", err)
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

	log.Printf("Creating album %s", title)
	album, err := p.service.CreateAlbum(ctx, &photoslibrary.CreateAlbumRequest{
		Album: &photoslibrary.Album{Title: title},
	})
	if err != nil {
		return nil, fmt.Errorf("Could not create an album: %s", err)
	}

	for _, chunk := range splitMediaItems(newMediaItems, batchCreateSize) {
		log.Printf("Adding %d item(s) into the album %s", len(chunk), album.Title)
		batch := &photoslibrary.BatchCreateMediaItemsRequest{
			NewMediaItems: chunk,
			AlbumId:       album.Id,
			AlbumPosition: &photoslibrary.AlbumPosition{Position: "LAST_IN_ALBUM"},
		}
		if err := p.service.BatchCreate(ctx, batch); err != nil {
			return nil, fmt.Errorf("Could not add items to the album: %s", err)
		}
	}
	return album, nil
}
