package internal

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/lestrrat-go/backoff"
)

type uploadService interface {
	Upload(context.Context, UploadItem) (UploadToken, error)
}

// UploadItem represents an uploadable item.
type UploadItem interface {
	// Open returns a stream.
	// Caller should close it finally.
	Open() (io.ReadCloser, int64, error)
	// Name returns the filename.
	Name() string
	// String returns the full name, e.g. path or URL.
	String() string
}

// UploadToken represents a pointer to the uploaded item.
type UploadToken string

const uploadEndpoint = "https://photoslibrary.googleapis.com/v1/uploads"

// Upload uploads the media item.
// It returns an upload token. You can append it to the library by `Append()`.
// It will retry uploading if status code is 5xx or network error occurs.
// See https://developers.google.com/photos/library/guides/best-practices#retrying-failed-requests
func (p *defaultPhotos) Upload(ctx context.Context, uploadItem UploadItem) (UploadToken, error) {
	b, cancel := defaultRetryPolicy.Start(ctx)
	defer cancel()
	for backoff.Continue(b) {
		r, size, err := uploadItem.Open()
		if err != nil {
			return "", fmt.Errorf("Could not open %s: %s", uploadItem, err)
		}
		defer r.Close()

		req, err := http.NewRequest("POST", uploadEndpoint, r)
		if err != nil {
			return "", fmt.Errorf("Could not create a request for uploading %s: %s", uploadItem, err)
		}
		req = req.WithContext(ctx)
		req.ContentLength = size
		req.Header.Add("Content-Type", "application/octet-stream")
		req.Header.Add("X-Goog-Upload-File-Name", uploadItem.Name())
		req.Header.Add("X-Goog-Upload-Protocol", "raw")

		p.log.Printf("Uploading %s (%d kB)", uploadItem.Name(), size/1024)
		res, err := p.client.Do(req)
		if err != nil {
			p.log.Printf("Error while uploading %s: %s", uploadItem, err)
			continue
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			p.log.Printf("Error while uploading %s: %s: could not read body: %s", uploadItem, res.Status, err)
			continue
		}
		body := string(b)

		switch {
		case res.StatusCode == 200:
			return UploadToken(body), nil
		case IsRetryableStatusCode(res.StatusCode):
			p.log.Printf("Error while uploading %s: %s: %s", uploadItem, res.Status, body)
		default:
			return "", fmt.Errorf("Got %s: %s", res.Status, body)
		}
	}
	return "", fmt.Errorf("Retry over")
}
