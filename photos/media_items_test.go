package photos

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFileMediaItem(t *testing.T) {
	tempdir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatalf("Could not create a temporary directory: %s", err)
	}
	defer os.RemoveAll(tempdir)
	if err := ioutil.WriteFile(tempdir+"/foo.jpg", []byte("example"), 0644); err != nil {
		t.Fatalf("Could not write bytes to file: %s", err)
	}
	filename := tempdir + "/foo.jpg"
	item := FileMediaItem(filename)

	if "foo.jpg" != item.Name() {
		t.Errorf("Name() wants %s but %s", "foo.jpg", item.Name())
	}
	if filename != item.String() {
		t.Errorf("Name() wants %s but %s", filename, item.String())
	}
	r, l, err := item.Open()
	if err != nil {
		t.Errorf("Open() returns error: %s", err)
	}
	defer r.Close()
	if 7 != l {
		t.Errorf("Content length wants 7 but %d", l)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("Could not read from Open(): %s", err)
	}
	if "example" != string(b) {
		t.Errorf("Content wants example but %s", string(b))
	}
}

func TestHTTPMediaItem(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/foo.jpg":
			w.Write([]byte("example"))
		default:
			http.Error(w, "Not Found", 404)
		}
	}))
	defer s.Close()
	url := s.URL + "/foo.jpg"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Could not create a request: %s", err)
	}
	item := &HTTPMediaItem{
		Client:  http.DefaultClient,
		Request: req,
	}

	if "foo.jpg" != item.Name() {
		t.Errorf("Name() wants %s but %s", "foo.jpg", item.Name())
	}
	if url != item.String() {
		t.Errorf("Name() wants %s but %s", url, item.String())
	}
	r, l, err := item.Open()
	if err != nil {
		t.Errorf("Open() returns error: %s", err)
	}
	defer r.Close()
	if 7 != l {
		t.Errorf("Content length wants 7 but %d", l)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("Could not read from Open(): %s", err)
	}
	if "example" != string(b) {
		t.Errorf("Content wants example but %s", string(b))
	}
}
