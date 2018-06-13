package photos

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"

	"google.golang.org/api/photoslibrary/v1"
)

const batchCreateSize = 50

const uploadConcurrency = 3

const apiVersion = "v1"
const basePath = "https://photoslibrary.googleapis.com/"

// Photos provides service for manage albums and uploading files.
type Photos struct {
	client  *http.Client
	service *photoslibrary.Service
	log     *log.Logger
}

// New creates a Photos.
func New(client *http.Client) (*Photos, error) {
	service, err := photoslibrary.New(client)
	if err != nil {
		return nil, err
	}
	return &Photos{
		client:  client,
		service: service,
		log:     log.New(os.Stdout, "", log.LstdFlags),
	}, nil
}

// AddToLibrary adds the files to the library.
// This method tries uploading all files and ignores any error.
// If no file could be uploaded, this method returns an error.
func (p *Photos) AddToLibrary(filepaths []string) error {
	mediaItems := p.UploadFiles(filepaths)
	if len(mediaItems) == 0 {
		return fmt.Errorf("Could not upload any file")
	}
	for _, chunk := range splitMediaItems(mediaItems, batchCreateSize) {
		p.log.Printf("Adding %d file(s) to the library", len(chunk))
		_, err := p.service.MediaItems.BatchCreate(&photoslibrary.BatchCreateMediaItemsRequest{
			NewMediaItems: chunk,
		}).Do()
		if err != nil {
			return fmt.Errorf("Error while adding files to the album: %s", err)
		}
	}
	return nil
}

// CreateAlbum creates an album with the files.
// This method tries uploading all files and ignores any error.
// If no file could be uploaded, this method returns an error.
func (p *Photos) CreateAlbum(title string, filepaths []string) (*photoslibrary.Album, error) {
	mediaItems := p.UploadFiles(filepaths)
	if len(mediaItems) == 0 {
		return nil, fmt.Errorf("Could not upload any file")
	}

	p.log.Printf("Creating album %s", title)
	album, err := p.service.Albums.Create(&photoslibrary.CreateAlbumRequest{
		Album: &photoslibrary.Album{
			Title: title,
		},
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("Error while creating an album: %s", err)
	}

	for _, chunk := range splitMediaItems(mediaItems, batchCreateSize) {
		p.log.Printf("Adding %d file(s) into the album %s", len(chunk), album.Title)
		_, err := p.service.MediaItems.BatchCreate(&photoslibrary.BatchCreateMediaItemsRequest{
			AlbumId:       album.Id,
			NewMediaItems: chunk,
		}).Do()
		if err != nil {
			return nil, fmt.Errorf("Error while adding files to the album: %s", err)
		}
	}
	return album, nil
}

// UploadFiles uploads the files.
// This method tries uploading all files and ignores any error.
// If no file could be uploaded, this method returns an empty array.
func (p *Photos) UploadFiles(filepaths []string) []*photoslibrary.NewMediaItem {
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
		go p.uploadWorker(uploadQueue, aggregateQueue, workerGroup)
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

func (p *Photos) uploadWorker(uploadQueue chan string, aggregateQueue chan *photoslibrary.NewMediaItem, workerGroup *sync.WaitGroup) {
	defer workerGroup.Done()
	for filepath := range uploadQueue {
		mediaItem, err := p.UploadFile(filepath)
		if err != nil {
			p.log.Printf("Error while uploading file %s: %s", filepath, err)
		} else {
			aggregateQueue <- mediaItem
		}
	}
}

// UploadFile uploads the file.
func (p *Photos) UploadFile(filepath string) (*photoslibrary.NewMediaItem, error) {
	r, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("Could not open file %s: %s", filepath, err)
	}
	defer r.Close()

	filename := path.Base(filepath)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s/uploads", basePath, apiVersion), r)
	if err != nil {
		return nil, fmt.Errorf("Could not create a request for uploading file %s: %s", filepath, err)
	}
	req.Header.Add("X-Goog-Upload-File-Name", filename)

	p.log.Printf("Uploading %s", filepath)
	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Could not send a request for uploading file %s: %s", filepath, err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Could not read the response body while uploading file %s: status=%d, %s", filepath, res.StatusCode, err)
	}
	body := string(b)

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Could not upload file %s: status=%d, body=%s", filepath, res.StatusCode, body)
	}
	return &photoslibrary.NewMediaItem{
		Description: filename,
		SimpleMediaItem: &photoslibrary.SimpleMediaItem{
			UploadToken: body,
		},
	}, nil
}
