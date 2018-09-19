package photos

import (
	"context"
	"log"
	"sync"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

const uploadConcurrency = 4

// UploadMediaItems uploads the media items.
// This method tries uploading all items and ignores any error.
// If no file could be uploaded, this method returns an empty array.
func (p *Photos) UploadMediaItems(ctx context.Context, mediaItems []MediaItem) []*photoslibrary.NewMediaItem {
	uploadQueue := make(chan MediaItem, len(mediaItems))
	for _, mediaItem := range mediaItems {
		uploadQueue <- mediaItem
	}
	close(uploadQueue)
	log.Printf("Queued %d item(s)", len(mediaItems))

	aggregateQueue := make(chan *photoslibrary.NewMediaItem, len(mediaItems))
	workerGroup := new(sync.WaitGroup)
	for i := 0; i < uploadConcurrency; i++ {
		workerGroup.Add(1)
		go p.uploadWorker(ctx, uploadQueue, aggregateQueue, workerGroup)
	}
	go func() {
		workerGroup.Wait()
		close(aggregateQueue)
	}()

	newMediaItems := make([]*photoslibrary.NewMediaItem, 0, len(mediaItems))
	for mediaItem := range aggregateQueue {
		newMediaItems = append(newMediaItems, mediaItem)
	}
	return newMediaItems
}

func (p *Photos) uploadWorker(ctx context.Context, uploadQueue chan MediaItem, aggregateQueue chan *photoslibrary.NewMediaItem, workerGroup *sync.WaitGroup) {
	defer workerGroup.Done()
	for mediaItem := range uploadQueue {
		newMediaItem, err := p.service.UploadMediaItem(ctx, mediaItem)
		if err != nil {
			log.Printf("Error while uploading %s: %s", mediaItem, err)
		} else {
			aggregateQueue <- newMediaItem
		}
	}
}
