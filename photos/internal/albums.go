package internal

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/backoff"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

type albumsService interface {
	CreateAlbum(context.Context, *photoslibrary.CreateAlbumRequest) (*photoslibrary.Album, error)
	ListAlbums(ctx context.Context, pageSize int64, pageToken string) (*photoslibrary.ListAlbumsResponse, error)
}

func (p *defaultPhotos) CreateAlbum(ctx context.Context, req *photoslibrary.CreateAlbumRequest) (*photoslibrary.Album, error) {
	create := p.service.Albums.Create(req)
	b, cancel := defaultRetryPolicy.Start(ctx)
	defer cancel()
	for backoff.Continue(b) {
		res, err := create.Do()
		switch {
		case err == nil:
			return res, nil
		case IsRetryableError(err):
			p.log.Printf("Error while creating an album: %s", err)
		default:
			return nil, err
		}
	}
	return nil, fmt.Errorf("Retry over")
}

func (p *defaultPhotos) ListAlbums(ctx context.Context, pageSize int64, pageToken string) (*photoslibrary.ListAlbumsResponse, error) {
	list := p.service.Albums.List().PageSize(pageSize).PageToken(pageToken)
	b, cancel := defaultRetryPolicy.Start(ctx)
	defer cancel()
	for backoff.Continue(b) {
		res, err := list.Do()
		switch {
		case err == nil:
			return res, nil
		case IsRetryableError(err):
			p.log.Printf("Error while listing albums: %s", err)
		default:
			return nil, err
		}
	}
	return nil, fmt.Errorf("Retry over")
}
