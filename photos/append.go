package photos

import (
	"context"
	"fmt"
	"time"

	"github.com/lestrrat-go/backoff"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

var appendRetryPolicy = backoff.NewExponential(
	backoff.WithInterval(3*time.Second),
	backoff.WithMaxRetries(5),
)

// Append appends the items to the album or your library (if albumId is empty).
// If some item(s) have been failed, this method does not return an error but prints message(s).
// If a network error occurs, this method retries and finally returns the error.
func (p *Photos) Append(ctx context.Context, albumID string, mediaItems []*photoslibrary.NewMediaItem) error {
	batch := p.service.MediaItems.BatchCreate(&photoslibrary.BatchCreateMediaItemsRequest{
		NewMediaItems: mediaItems,
		AlbumId:       albumID,
		AlbumPosition: &photoslibrary.AlbumPosition{Position: "LAST_IN_ALBUM"},
	})
	b, cancel := appendRetryPolicy.Start(ctx)
	defer cancel()
	for backoff.Continue(b) {
		res, err := batch.Do()
		switch {
		case err == nil:
			for _, result := range res.NewMediaItemResults {
				if result.Status.Code != 0 {
					if mediaItem := findMediaItemByUploadToken(mediaItems, result.UploadToken); mediaItem != nil {
						p.log.Printf("Skipped %s: %s (%d)", mediaItem.Description, result.Status.Message, result.Status.Code)
					} else {
						p.log.Printf("Error while adding files: %s (%d)", result.Status.Message, result.Status.Code)
					}
				}
			}
			return nil
		case IsRetryableError(err):
			p.log.Printf("Error while BatchCreate: %s", err)
		default:
			return fmt.Errorf("Error while BatchCreate: %s", err)
		}
	}
	return fmt.Errorf("Could not perform BatchCreate: retry over")
}

func findMediaItemByUploadToken(mediaItems []*photoslibrary.NewMediaItem, uploadToken string) *photoslibrary.NewMediaItem {
	for _, mediaItem := range mediaItems {
		if mediaItem.SimpleMediaItem.UploadToken == uploadToken {
			return mediaItem
		}
	}
	return nil
}
