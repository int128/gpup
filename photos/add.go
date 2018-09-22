package photos

import (
	"context"
	"fmt"
	"log"
	"sync"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

const uploadConcurrency = 4
const batchCreateSize = 10

// AddToLibrary adds the items to the library.
// This method tries uploading all items and ignores any error.
// If no item could be uploaded, this method returns an error.
func (p *Photos) AddToLibrary(ctx context.Context, mediaItems []MediaItem) error {
	return p.addMediaItems(ctx, mediaItems, &photoslibrary.BatchCreateMediaItemsRequest{})
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
	return p.addMediaItems(ctx, mediaItems, &photoslibrary.BatchCreateMediaItemsRequest{
		AlbumId:       album.Id,
		AlbumPosition: &photoslibrary.AlbumPosition{Position: "LAST_IN_ALBUM"},
	})
}

// CreateAlbum creates an album with the media items.
// This method tries uploading all items and ignores any error.
// If no item could be uploaded, this method returns an error.
func (p *Photos) CreateAlbum(ctx context.Context, title string, mediaItems []MediaItem) error {
	log.Printf("Creating album %s", title)
	album, err := p.service.CreateAlbum(ctx, &photoslibrary.CreateAlbumRequest{
		Album: &photoslibrary.Album{Title: title},
	})
	if err != nil {
		return fmt.Errorf("Could not create an album: %s", err)
	}
	return p.addMediaItems(ctx, mediaItems, &photoslibrary.BatchCreateMediaItemsRequest{
		AlbumId:       album.Id,
		AlbumPosition: &photoslibrary.AlbumPosition{Position: "LAST_IN_ALBUM"},
	})
}

func (p *Photos) addMediaItems(ctx context.Context, mediaItems []MediaItem, batchCreateRequest *photoslibrary.BatchCreateMediaItemsRequest) error {
	uploadQueue := make(chan MediaItem, len(mediaItems))
	for _, mediaItem := range mediaItems {
		uploadQueue <- mediaItem
	}
	close(uploadQueue)
	log.Printf("Queued %d item(s)", len(mediaItems))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	batchCreateQueue := make(chan *photoslibrary.NewMediaItem, len(mediaItems))
	var wg sync.WaitGroup
	for i := 0; i < uploadConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for mediaItem := range uploadQueue {
				newMediaItem, err := p.service.UploadMediaItem(ctx, mediaItem)
				if err != nil {
					log.Printf("Error while uploading %s: %s", mediaItem, err)
				} else {
					batchCreateQueue <- newMediaItem
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(batchCreateQueue)
	}()

	buffer := batchCreateBuffer{
		Size: batchCreateSize,
		Trigger: func(newMediaItems []*photoslibrary.NewMediaItem) error {
			log.Printf("Adding %d item(s)", len(newMediaItems))
			if err := p.service.BatchCreate(ctx, &photoslibrary.BatchCreateMediaItemsRequest{
				NewMediaItems: newMediaItems,
				AlbumId:       batchCreateRequest.AlbumId,
				AlbumPosition: batchCreateRequest.AlbumPosition,
			}); err != nil {
				return fmt.Errorf("Could not add items: %s", err)
			}
			return nil
		},
	}
	for newMediaItem := range batchCreateQueue {
		if err := buffer.Add(newMediaItem); err != nil {
			return err
		}
	}
	return buffer.Flush()
}

type batchCreateBuffer struct {
	Size    int
	Trigger func([]*photoslibrary.NewMediaItem) error

	batch []*photoslibrary.NewMediaItem
}

func (b *batchCreateBuffer) empty() {
	b.batch = make([]*photoslibrary.NewMediaItem, 0)
}

func (b *batchCreateBuffer) Add(item *photoslibrary.NewMediaItem) error {
	if b.batch == nil {
		b.empty()
	}
	b.batch = append(b.batch, item)
	if len(b.batch) >= b.Size {
		defer b.empty()
		return b.Trigger(b.batch)
	}
	return nil
}

func (b *batchCreateBuffer) Flush() error {
	if len(b.batch) > 0 {
		defer b.empty()
		return b.Trigger(b.batch)
	}
	return nil
}
