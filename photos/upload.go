package photos

import (
	"context"
	"fmt"
	"io/ioutil"
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

// UploadFiles uploads the files.
// This method tries uploading all files and ignores any error.
// If no file could be uploaded, this method returns an empty array.
func (p *Photos) UploadFiles(ctx context.Context, filepaths []string) []*photoslibrary.NewMediaItem {
	uploadQueue := make(chan string, len(filepaths))
	for _, filepath := range filepaths {
		uploadQueue <- filepath
	}
	close(uploadQueue)
	p.log.Printf("Queued %d file(s)", len(filepaths))

	aggregateQueue := make(chan *photoslibrary.NewMediaItem, len(filepaths))
	workerGroup := new(sync.WaitGroup)
	for i := 0; i < uploadConcurrency; i++ {
		workerGroup.Add(1)
		go p.uploadWorker(ctx, uploadQueue, aggregateQueue, workerGroup)
	}
	go func() {
		workerGroup.Wait()
		close(aggregateQueue)
	}()

	mediaItems := make([]*photoslibrary.NewMediaItem, 0, len(filepaths))
	for mediaItem := range aggregateQueue {
		mediaItems = append(mediaItems, mediaItem)
	}
	return mediaItems
}

func (p *Photos) uploadWorker(ctx context.Context, uploadQueue chan string, aggregateQueue chan *photoslibrary.NewMediaItem, workerGroup *sync.WaitGroup) {
	defer workerGroup.Done()
	for filepath := range uploadQueue {
		mediaItem, err := p.UploadFile(ctx, filepath)
		if err != nil {
			p.log.Printf("Error while uploading file %s: %s", filepath, err)
		} else {
			aggregateQueue <- mediaItem
		}
	}
}

// UploadFile uploads the file.
// It returns an upload token. You can append it to the library by `Append()`.
// It will retry uploading if status code is 5xx or network error occurs.
// See https://developers.google.com/photos/library/guides/best-practices#retrying-failed-requests
func (p *Photos) UploadFile(ctx context.Context, filepath string) (*photoslibrary.NewMediaItem, error) {
	filename := path.Base(filepath)
	b, cancel := uploadRetryPolicy.Start(ctx)
	defer cancel()
	for backoff.Continue(b) {
		r, err := os.Open(filepath)
		if err != nil {
			return nil, fmt.Errorf("Could not open file %s: %s", filepath, err)
		}
		defer r.Close()

		req, err := http.NewRequest("POST", fmt.Sprintf("%s%s/uploads", basePath, apiVersion), r)
		if err != nil {
			return nil, fmt.Errorf("Could not create a request for uploading file %s: %s", filepath, err)
		}
		req.Header.Add("X-Goog-Upload-File-Name", filename)
		req.Header.Add("X-Goog-Upload-Protocol", "raw")

		p.log.Printf("Uploading %s", filepath)
		res, err := p.client.Do(req)
		if err != nil {
			p.log.Printf("Error while uploading %s: %s", filepath, err)
			continue
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			p.log.Printf("Error while uploading %s: status %d: could not read body: %s", filepath, res.StatusCode, err)
			continue
		}
		body := string(b)

		switch {
		case res.StatusCode == 200:
			return &photoslibrary.NewMediaItem{
				Description:     filename,
				SimpleMediaItem: &photoslibrary.SimpleMediaItem{UploadToken: body},
			}, nil
		case IsRetryableStatusCode(res.StatusCode):
			p.log.Printf("Error while uploading %s: status %d: %s", filepath, res.StatusCode, body)
		default:
			return nil, fmt.Errorf("Could not upload %s: status %d: %s", filepath, res.StatusCode, body)
		}
	}
	return nil, fmt.Errorf("Could not upload %s: retry over", filepath)
}
