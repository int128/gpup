package photos

import (
	"context"
	"fmt"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

// ListAlbumsFunc is called for each response of 50 albums.
// If this calls stop, ListAlbums stops the loop.
type ListAlbumsFunc func(albums []*photoslibrary.Album, stop func())

// ListAlbums gets a list of albums.
// It calls the function for each 50 albums.
func (p *Photos) ListAlbums(ctx context.Context, callback ListAlbumsFunc) error {
	var nextPageToken string
	for {
		albums, err := p.service.Albums.List().PageSize(50).PageToken(nextPageToken).Do()
		if err != nil {
			return fmt.Errorf("Error while listing albums: %s", err)
		}
		var stop bool
		callback(albums.Albums, func() { stop = true })
		if stop {
			return nil
		}
		nextPageToken = albums.NextPageToken
		if nextPageToken == "" {
			return nil
		}
	}
}

// FindAlbumByTitle returns the album which has the title.
// If the album was not found, it returns nil.
// If any error occurred, it returns the error.
func (p *Photos) FindAlbumByTitle(ctx context.Context, title string) (*photoslibrary.Album, error) {
	var matched *photoslibrary.Album
	if err := p.ListAlbums(ctx, func(albums []*photoslibrary.Album, stop func()) {
		for _, album := range albums {
			if album.Title == title {
				stop()
				matched = album
				return
			}
		}
	}); err != nil {
		return nil, err
	}
	return matched, nil
}
