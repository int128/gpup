package photos

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/int128/gpup/photos/internal"
)

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
