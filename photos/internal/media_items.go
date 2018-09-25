package internal

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/backoff"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

type mediaItemsService interface {
	BatchCreate(context.Context, *photoslibrary.BatchCreateMediaItemsRequest) (*photoslibrary.BatchCreateMediaItemsResponse, error)
}

// BatchCreate creates the items to the album or your library.
// If a network error occurs, this method retries and finally returns the error.
func (p *defaultPhotos) BatchCreate(ctx context.Context, req *photoslibrary.BatchCreateMediaItemsRequest) (*photoslibrary.BatchCreateMediaItemsResponse, error) {
	batch := p.service.MediaItems.BatchCreate(req)
	b, cancel := defaultRetryPolicy.Start(ctx)
	defer cancel()
	for backoff.Continue(b) {
		res, err := batch.Do()
		switch {
		case err == nil:
			return res, nil
		case IsRetryableError(err):
			p.log.Printf("Error while adding the item: %s", err)
		default:
			return nil, err
		}
	}
	return nil, fmt.Errorf("Retry over")
}
