package photos

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync/atomic"
	"testing"

	"github.com/int128/gpup/photos/internal"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

type serviceMock struct {
	uploadErrorFunc       func(internal.UploadItem) error
	uploadCalls           int32
	batchCreateErrorFunc  func(*photoslibrary.BatchCreateMediaItemsRequest) error
	batchCreateStatusFunc func(*photoslibrary.NewMediaItem) *photoslibrary.Status
	batchCreateCalls      []*photoslibrary.BatchCreateMediaItemsRequest
}

func (m *serviceMock) Upload(ctx context.Context, u internal.UploadItem) (internal.UploadToken, error) {
	log.Printf("Upload(%s)", u)
	atomic.AddInt32(&m.uploadCalls, 1)
	if m.uploadErrorFunc != nil {
		if err := m.uploadErrorFunc(u); err != nil {
			return "", err
		}
	}
	return internal.UploadToken(u.String()), nil
}

func (m *serviceMock) BatchCreate(ctx context.Context, r *photoslibrary.BatchCreateMediaItemsRequest) (*photoslibrary.BatchCreateMediaItemsResponse, error) {
	log.Printf("BatchCreate(%d)", len(r.NewMediaItems))
	m.batchCreateCalls = append(m.batchCreateCalls, r)
	if len(r.NewMediaItems) > batchCreateSize {
		return nil, fmt.Errorf("Got [%d]NewMediaItems over batchCreateSize=%d", len(r.NewMediaItems), batchCreateSize)
	}
	if m.batchCreateErrorFunc != nil {
		if err := m.batchCreateErrorFunc(r); err != nil {
			return nil, err
		}
	}
	results := make([]*photoslibrary.NewMediaItemResult, len(r.NewMediaItems))
	for i, item := range r.NewMediaItems {
		s := &photoslibrary.Status{Code: 0, Message: "OK"}
		if m.batchCreateStatusFunc != nil {
			s = m.batchCreateStatusFunc(item)
		}
		results[i] = &photoslibrary.NewMediaItemResult{
			Status:      s,
			UploadToken: item.SimpleMediaItem.UploadToken,
			MediaItem:   &photoslibrary.MediaItem{Description: item.SimpleMediaItem.UploadToken},
		}
	}
	return &photoslibrary.BatchCreateMediaItemsResponse{NewMediaItemResults: results}, nil
}

func (m *serviceMock) CreateAlbum(context.Context, *photoslibrary.CreateAlbumRequest) (*photoslibrary.Album, error) {
	return nil, fmt.Errorf("CreateAlbum not implemented")
}

func (m *serviceMock) ListAlbums(context.Context, int64, string) (*photoslibrary.ListAlbumsResponse, error) {
	return nil, fmt.Errorf("ListAlbums not implemented")
}

type uploadItemMock int

func (m uploadItemMock) Open() (io.ReadCloser, int64, error) {
	b := []byte(m.String())
	return ioutil.NopCloser(bytes.NewReader(b)), int64(len(b)), nil
}

func (m uploadItemMock) Name() string { return m.String() }

func (m uploadItemMock) String() string { return fmt.Sprintf("UploadItem#%d", m) }

func (m uploadItemMock) SizeAfterOpen() bool { return false }

func makeUploadItems(n int) []UploadItem {
	ret := make([]UploadItem, n)
	for i := 0; i < n; i++ {
		ret[i] = uploadItemMock(i)
	}
	return ret
}

func TestPhotos_add(t *testing.T) {
	defer func(restore int) { batchCreateSize = restore }(batchCreateSize)
	batchCreateSize = 10

	for _, c := range []struct {
		count            int
		batchCreateCalls int
	}{
		{9, 1},
		{10, 1},
		{11, 2},
		{19, 2},
		{20, 2},
		{21, 3},
	} {
		t.Run(fmt.Sprintf("count=%d", c.count), func(t *testing.T) {
			m := &serviceMock{}
			p := &Photos{service: m}
			uploadItems := makeUploadItems(c.count)
			results := p.add(context.Background(), uploadItems, photoslibrary.BatchCreateMediaItemsRequest{})
			if len(results) != len(uploadItems) {
				t.Errorf("len(results) wants %d but %d", len(uploadItems), len(results))
			}
			for i, r := range results {
				if r.Error != nil {
					t.Errorf("r[%d].Error wants nil but %s", i, r.Error)
				}
				if r.MediaItem == nil {
					t.Errorf("r[%d].MediaItem wants non-nil but nil", i)
				}
			}
			if int(m.uploadCalls) != len(uploadItems) {
				t.Errorf("Upload API call wants %d times but %d", len(uploadItems), m.uploadCalls)
			}
			if len(m.batchCreateCalls) != c.batchCreateCalls {
				t.Errorf("BatchCreate API call wants %d times but %d", c.batchCreateCalls, len(m.batchCreateCalls))
			}
		})
	}
}

func TestPhotos_add_error(t *testing.T) {
	defer func(restore int) { batchCreateSize = restore }(batchCreateSize)
	batchCreateSize = 10

	uploadError := serviceMock{
		uploadErrorFunc: func(internal.UploadItem) error { return fmt.Errorf("ERR") },
	}
	batchCreateError := serviceMock{
		batchCreateErrorFunc: func(*photoslibrary.BatchCreateMediaItemsRequest) error { return fmt.Errorf("ERR") },
	}
	batchCreateStatus1 := serviceMock{
		batchCreateStatusFunc: func(*photoslibrary.NewMediaItem) *photoslibrary.Status {
			return &photoslibrary.Status{Code: 1, Message: "ERR"}
		},
	}

	for _, c := range []struct {
		count            int
		name             string
		m                serviceMock
		batchCreateCalls int
	}{
		{9, "uploadError", uploadError, 0},
		{10, "uploadError", uploadError, 0},
		{11, "uploadError", uploadError, 0},
		{9, "batchCreateError", batchCreateError, 1},
		{10, "batchCreateError", batchCreateError, 1},
		{11, "batchCreateError", batchCreateError, 2},
		{9, "batchCreateStatus1", batchCreateStatus1, 1},
		{10, "batchCreateStatus1", batchCreateStatus1, 1},
		{11, "batchCreateStatus1", batchCreateStatus1, 2},
	} {
		t.Run(fmt.Sprintf("count=%d/%s", c.count, c.name), func(t *testing.T) {
			p := &Photos{service: &c.m}
			uploadItems := makeUploadItems(c.count)
			results := p.add(context.Background(), uploadItems, photoslibrary.BatchCreateMediaItemsRequest{})
			if len(results) != len(uploadItems) {
				t.Errorf("len(results) wants %d but %d", len(uploadItems), len(results))
			}
			for i, r := range results {
				if r.Error == nil {
					t.Errorf("r[%d].Error wants non-nil but nil", i)
				}
				if r.MediaItem != nil {
					t.Errorf("r[%d].MediaItem wants nil but %+v", i, r.MediaItem)
				}
			}
			if int(c.m.uploadCalls) != len(uploadItems) {
				t.Errorf("Upload API call wants %d times but %d", len(uploadItems), c.m.uploadCalls)
			}
			if len(c.m.batchCreateCalls) != c.batchCreateCalls {
				t.Errorf("BatchCreate API call wants %d times but %d", c.batchCreateCalls, len(c.m.batchCreateCalls))
			}
		})
	}
}
