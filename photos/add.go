package photos

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/int128/gpup/photos/internal"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

const uploadConcurrency = 4
const batchCreateSize = 10

// AddToLibrary adds the items to the library.
// This method tries uploading all items and ignores any error.
// If no item could be uploaded, this method returns an error.
func (p *Photos) AddToLibrary(ctx context.Context, uploadItems []UploadItem) error {
	return p.addUploadItems(ctx, uploadItems, &photoslibrary.BatchCreateMediaItemsRequest{})
}

// AddToAlbum adds the items to the album.
// This method tries uploading all items and ignores any error.
// If no item could be uploaded, this method returns an error.
func (p *Photos) AddToAlbum(ctx context.Context, title string, uploadItems []UploadItem) error {
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
	return p.addUploadItems(ctx, uploadItems, &photoslibrary.BatchCreateMediaItemsRequest{
		AlbumId:       album.Id,
		AlbumPosition: &photoslibrary.AlbumPosition{Position: "LAST_IN_ALBUM"},
	})
}

// CreateAlbum creates an album with the media items.
// This method tries uploading all items and ignores any error.
// If no item could be uploaded, this method returns an error.
func (p *Photos) CreateAlbum(ctx context.Context, title string, uploadItems []UploadItem) error {
	log.Printf("Creating album %s", title)
	album, err := p.service.CreateAlbum(ctx, &photoslibrary.CreateAlbumRequest{
		Album: &photoslibrary.Album{Title: title},
	})
	if err != nil {
		return fmt.Errorf("Could not create an album: %s", err)
	}
	return p.addUploadItems(ctx, uploadItems, &photoslibrary.BatchCreateMediaItemsRequest{
		AlbumId:       album.Id,
		AlbumPosition: &photoslibrary.AlbumPosition{Position: "LAST_IN_ALBUM"},
	})
}

func (p *Photos) addUploadItems(ctx context.Context, uploadItems []UploadItem, batchCreateRequest *photoslibrary.BatchCreateMediaItemsRequest) error {
	uploadQueue := make(chan UploadItem, len(uploadItems))
	for _, uploadItem := range uploadItems {
		uploadQueue <- uploadItem
	}
	close(uploadQueue)
	log.Printf("Queued %d item(s)", len(uploadItems))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	batchCreateQueue := make(chan internal.UploadToken, len(uploadItems))
	var wg sync.WaitGroup
	for i := 0; i < uploadConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for uploadItem := range uploadQueue {
				uploadToken, err := p.service.Upload(ctx, uploadItem)
				if err != nil {
					log.Printf("Error while uploading %s: %s", uploadItem, err)
				} else {
					batchCreateQueue <- uploadToken
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
			res, err := p.service.BatchCreate(ctx, &photoslibrary.BatchCreateMediaItemsRequest{
				NewMediaItems: newMediaItems,
				AlbumId:       batchCreateRequest.AlbumId,
				AlbumPosition: batchCreateRequest.AlbumPosition,
			})
			if err != nil {
				return fmt.Errorf("Could not add items: %s", err)
			}
			for _, result := range res.NewMediaItemResults {
				if result.Status.Code != 0 {
					if mediaItem := findMediaItemByUploadToken(newMediaItems, result.UploadToken); mediaItem != nil {
						log.Printf("Skipped %s: %s (%d)", mediaItem.Description, result.Status.Message, result.Status.Code)
					} else {
						log.Printf("Error while adding the item: %s (%d)", result.Status.Message, result.Status.Code)
					}
				}
			}
			return nil
		},
	}
	for uploadToken := range batchCreateQueue {
		if err := buffer.Add(&photoslibrary.NewMediaItem{
			SimpleMediaItem: &photoslibrary.SimpleMediaItem{UploadToken: string(uploadToken)},
		}); err != nil {
			return err
		}
	}
	return buffer.Flush()
}

func findMediaItemByUploadToken(mediaItems []*photoslibrary.NewMediaItem, uploadToken string) *photoslibrary.NewMediaItem {
	for _, mediaItem := range mediaItems {
		if mediaItem.SimpleMediaItem.UploadToken == uploadToken {
			return mediaItem
		}
	}
	return nil
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
