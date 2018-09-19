package internal

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/lestrrat-go/backoff"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

type uploadService interface {
	UploadMediaItem(context.Context, MediaItem) (*photoslibrary.NewMediaItem, error)
}

// MediaItem represents an uploadable item.
type MediaItem interface {
	// Open returns a stream.
	// Caller should close it finally.
	Open() (io.ReadCloser, int64, error)
	// Name returns the filename.
	Name() string
	// String returns the full name, e.g. path or URL.
	String() string
}

const uploadEndpoint = "https://photoslibrary.googleapis.com/v1/uploads"

// UploadMediaItem uploads the media item.
// It returns an upload token. You can append it to the library by `Append()`.
// It will retry uploading if status code is 5xx or network error occurs.
// See https://developers.google.com/photos/library/guides/best-practices#retrying-failed-requests
func (p *defaultPhotos) UploadMediaItem(ctx context.Context, mediaItem MediaItem) (*photoslibrary.NewMediaItem, error) {
	b, cancel := defaultRetryPolicy.Start(ctx)
	defer cancel()
	for backoff.Continue(b) {
		r, size, err := mediaItem.Open()
		if err != nil {
			return nil, fmt.Errorf("Could not open %s: %s", mediaItem, err)
		}
		defer r.Close()

		req, err := http.NewRequest("POST", uploadEndpoint, r)
		if err != nil {
			return nil, fmt.Errorf("Could not create a request for uploading %s: %s", mediaItem, err)
		}
		req = req.WithContext(ctx)
		req.ContentLength = size
		req.Header.Add("Content-Type", "application/octet-stream")
		req.Header.Add("X-Goog-Upload-File-Name", mediaItem.Name())
		req.Header.Add("X-Goog-Upload-Protocol", "raw")

		p.log.Printf("Uploading %s (%d kB)", mediaItem.Name(), size/1024)
		res, err := p.client.Do(req)
		if err != nil {
			p.log.Printf("Error while uploading %s: %s", mediaItem, err)
			continue
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			p.log.Printf("Error while uploading %s: %s: could not read body: %s", mediaItem, res.Status, err)
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
			p.log.Printf("Error while uploading %s: %s: %s", mediaItem, res.Status, body)
		default:
			return nil, fmt.Errorf("Got %s: %s", res.Status, body)
		}
	}
	return nil, fmt.Errorf("Retry over")
}
