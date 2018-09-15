package photos

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/lestrrat-go/backoff"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

const uploadConcurrency = 4

var uploadRetryPolicy = backoff.NewExponential(
	backoff.WithInterval(3*time.Second),
	backoff.WithMaxRetries(5),
)

const apiVersion = "v1"
const basePath = "https://photoslibrary.googleapis.com/"

// MediaItem represents an uploadable item.
type MediaItem interface {
	// Open returns a stream.
	// Caller should close it finally.
	Open() (io.ReadCloser, error)
	// Name returns the filename.
	Name() string
	// String returns the full name, e.g. path or URL.
	String() string
}

// FileMediaItem represents a local file.
type FileMediaItem string

// Open returns a stream.
// Caller should close it finally.
func (m FileMediaItem) Open() (io.ReadCloser, error) {
	return os.Open(m.String())
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
func (m *HTTPMediaItem) Open() (io.ReadCloser, error) {
	r, err := m.Client.Do(m.Request)
	if err != nil {
		return nil, err
	}
	if r.StatusCode < 200 || r.StatusCode > 299 {
		r.Body.Close()
		return nil, fmt.Errorf("Got %s", r.Status)
	}
	log.Printf("%s %s", r.Status, m.Request.URL)
	return r.Body, nil
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
	p.log.Printf("Queued %d item(s)", len(mediaItems))

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
		newMediaItem, err := p.UploadMediaItem(ctx, mediaItem)
		if err != nil {
			p.log.Printf("Error while uploading %s: %s", mediaItem, err)
		} else {
			aggregateQueue <- newMediaItem
		}
	}
}

// UploadMediaItem uploads the media item.
// It returns an upload token. You can append it to the library by `Append()`.
// It will retry uploading if status code is 5xx or network error occurs.
// See https://developers.google.com/photos/library/guides/best-practices#retrying-failed-requests
func (p *Photos) UploadMediaItem(ctx context.Context, mediaItem MediaItem) (*photoslibrary.NewMediaItem, error) {
	b, cancel := uploadRetryPolicy.Start(ctx)
	defer cancel()
	for backoff.Continue(b) {
		r, err := mediaItem.Open()
		if err != nil {
			return nil, fmt.Errorf("Could not open %s: %s", mediaItem, err)
		}
		defer r.Close()

		req, err := http.NewRequest("POST", fmt.Sprintf("%s%s/uploads", basePath, apiVersion), r)
		if err != nil {
			return nil, fmt.Errorf("Could not create a request for uploading %s: %s", mediaItem, err)
		}
		req.Header.Add("X-Goog-Upload-File-Name", mediaItem.Name())
		req.Header.Add("X-Goog-Upload-Protocol", "raw")

		p.log.Printf("Uploading %s", mediaItem.Name())
		res, err := p.client.Do(req)
		if err != nil {
			p.log.Printf("Error while uploading %s: %s", mediaItem, err)
			continue
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			p.log.Printf("Error while uploading %s: status %d: could not read body: %s", mediaItem, res.StatusCode, err)
			continue
		}
		body := string(b)

		switch {
		case res.StatusCode == 200:
			return &photoslibrary.NewMediaItem{
				Description:     mediaItem.Name(),
				SimpleMediaItem: &photoslibrary.SimpleMediaItem{UploadToken: body},
			}, nil
		case IsRetryableStatusCode(res.StatusCode):
			p.log.Printf("Error while uploading %s: status %d: %s", mediaItem, res.StatusCode, body)
		default:
			return nil, fmt.Errorf("Could not upload %s: status %d: %s", mediaItem, res.StatusCode, body)
		}
	}
	return nil, fmt.Errorf("Could not upload %s: retry over", mediaItem)
}
