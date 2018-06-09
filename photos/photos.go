package photos

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"google.golang.org/api/photoslibrary/v1"
)

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
		log:     log.New(os.Stdout, "photos", log.LstdFlags),
	}, nil
}

// CreateAlbum creates an album with the files.
func (p *Photos) CreateAlbum(title string, filepaths []string) (*photoslibrary.Album, error) {
	p.log.Printf("Uploading %d files", len(filepaths))
	mediaItems, err := p.UploadFiles(filepaths)
	if err != nil {
		p.log.Printf("[warning] %s", err.Error())
		p.log.Printf("[warning] Continue to create an album with %d files", len(mediaItems))
	}

	p.log.Printf("Creating an album %s", title)
	album, err := p.service.Albums.Create(&photoslibrary.CreateAlbumRequest{
		Album: &photoslibrary.Album{
			Title: title,
		},
	}).Do()
	if err != nil {
		return nil, err
	}

	p.log.Printf("Adding %d files into the album id=%s", len(mediaItems), album.Id)
	batch, err := p.service.MediaItems.BatchCreate(&photoslibrary.BatchCreateMediaItemsRequest{
		AlbumId:       album.Id,
		NewMediaItems: mediaItems,
	}).Do()
	if err != nil {
		return nil, err
	}

	p.log.Printf("Added %d files into the album id=%s", len(batch.NewMediaItemResults), album.Id)
	return album, err
}

// UploadFiles uploads the files.
// If any error occurs while uploading, it continues uploading remaining and returns all errors.
func (p *Photos) UploadFiles(filepaths []string) ([]*photoslibrary.NewMediaItem, error) {
	items := make([]*photoslibrary.NewMediaItem, 0, len(filepaths))
	errs := make([]string, 0, len(filepaths))
	for _, filepath := range filepaths {
		item, err := p.UploadFile(filepath)
		if err != nil {
			errs = append(errs, err.Error())
		} else {
			items = append(items, item)
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("Could not upload some files:\n%s", strings.Join(errs, "\n"))
	}
	return items, nil
}

// UploadFile uploads the file.
func (p *Photos) UploadFile(filepath string) (*photoslibrary.NewMediaItem, error) {
	r, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("Could not open file %s: %s", filepath, err)
	}
	defer r.Close()

	filename := path.Base(filepath)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/uploads", basePath, apiVersion), r)
	if err != nil {
		return nil, fmt.Errorf("Could not create a request for uploading file %s: %s", filepath, err)
	}
	req.Header.Add("X-Goog-Upload-File-Name", filename)

	p.log.Printf("Uploading file %s", filepath)
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
