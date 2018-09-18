package photos

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/int128/gpup/photos/internal"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

const uploadConcurrency = 4

// MediaItem represents an uploadable item.
type MediaItem interface {
	internal.MediaItem
}

// FileMediaItem represents a local file.
type FileMediaItem string

// Open returns a stream.
// Caller should close it finally.
func (m FileMediaItem) Open() (io.ReadCloser, int64, error) {
	f, err := os.Stat(m.String())
	if err != nil {
		return nil, 0, err
	}
	r, err := os.Open(m.String())
	if err != nil {
		return nil, 0, err
	}
	return r, f.Size(), nil
}

// Name returns the filename.
func (m FileMediaItem) Name() string {
	return path.Base(m.String())
}

func (m FileMediaItem) String() string {
	return string(m)
}

// HTTPMediaItem represents a remote file.
type HTTPMediaItem struct {
	Client  *http.Client
	Request *http.Request
}

// Open returns a stream.
// Caller should close it finally.
func (m *HTTPMediaItem) Open() (io.ReadCloser, int64, error) {
	r, err := m.Client.Do(m.Request)
	if err != nil {
		return nil, 0, err
	}
	if r.StatusCode < 200 || r.StatusCode > 299 {
		r.Body.Close()
		return nil, 0, fmt.Errorf("Got %s", r.Status)
	}
	return r.Body, r.ContentLength, nil
}

// Name returns the filename.
func (m *HTTPMediaItem) Name() string {
	return path.Base(m.Request.URL.Path)
}

func (m *HTTPMediaItem) String() string {
	return m.Request.URL.String()
}

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
