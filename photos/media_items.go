package photos

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/int128/gpup/photos/internal"
)

// UploadItem represents an uploadable item.
type UploadItem interface {
	internal.UploadItem
}

// FileUploadItem represents a local file.
type FileUploadItem struct {
	name string
}

// NewFileUploadItem creates a new FileUploadItem
func NewFileUploadItem(name string) *FileUploadItem {
	return &FileUploadItem{
		name: name,
	}
}

// Open returns a stream.
// Caller should close it finally.
func (m FileUploadItem) Open() (io.ReadCloser, int64, error) {
	f, err := os.Stat(m.name)
	if err != nil {
		return nil, 0, err
	}
	r, err := os.Open(m.name)
	if err != nil {
		return nil, 0, err
	}
	return r, f.Size(), nil
}

// Name returns the filename.
func (m FileUploadItem) Name() string {
	return path.Base(m.name)
}

func (m FileUploadItem) String() string {
	return m.name
}

func (m FileUploadItem) SizeAfterOpen() bool {
	return false
}

// HTTPUploadItem represents a remote file.
type HTTPUploadItem struct {
	Client  *http.Client
	Request *http.Request
}

// Open returns a stream.
// Caller should close it finally.
func (m *HTTPUploadItem) Open() (io.ReadCloser, int64, error) {
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
func (m *HTTPUploadItem) Name() string {
	return path.Base(m.Request.URL.Path)
}

func (m *HTTPUploadItem) String() string {
	return m.Request.URL.String()
}

func (m HTTPUploadItem) SizeAfterOpen() bool {
	return true
}
