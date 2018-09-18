package internal

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/backoff"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

type mediaItemsService interface {
	BatchCreate(context.Context, *photoslibrary.BatchCreateMediaItemsRequest) error
}

// BatchCreate creates the items to the album or your library.
// If some item(s) have been failed, this method does not return an error but prints message(s).
// If a network error occurs, this method retries and finally returns the error.
func (p *defaultPhotos) BatchCreate(ctx context.Context, req *photoslibrary.BatchCreateMediaItemsRequest) error {
	batch := p.service.MediaItems.BatchCreate(req)
	b, cancel := defaultRetryPolicy.Start(ctx)
	defer cancel()
	for backoff.Continue(b) {
		res, err := batch.Do()
		switch {
		case err == nil:
			for _, result := range res.NewMediaItemResults {
				if result.Status.Code != 0 {
					if mediaItem := findMediaItemByUploadToken(req.NewMediaItems, result.UploadToken); mediaItem != nil {
						p.log.Printf("Skipped %s: %s (%d)", mediaItem.Description, result.Status.Message, result.Status.Code)
					} else {
						p.log.Printf("Error while adding the item: %s (%d)", result.Status.Message, result.Status.Code)
					}
				}
			}
			return nil
		case IsRetryableError(err):
			p.log.Printf("Error while adding the item: %s", err)
		default:
			return err
		}
	}
	return fmt.Errorf("Retry over")
}

func findMediaItemByUploadToken(mediaItems []*photoslibrary.NewMediaItem, uploadToken string) *photoslibrary.NewMediaItem {
	for _, mediaItem := range mediaItems {
		if mediaItem.SimpleMediaItem.UploadToken == uploadToken {
			return mediaItem
		}
	}
	return nil
}
