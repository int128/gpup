package cli

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/int128/gpup/photos"
)

func TestCLI_findMediaItems(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "FindFiles")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)
	if err := os.Chdir(tempdir); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("album1", 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile("album1/a.jpg", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("album2", 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile("album2/b.jpg", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile("album2/c.jpg", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	c := CLI{
		Paths: []string{
			".",
			"http://www.example.com/image.jpg",
		},
	}
	mediaItems, err := c.findMediaItems()
	if err != nil {
		t.Fatal(err)
	}
	if len(mediaItems) != 4 {
		t.Errorf("wants size 4 but %d", len(mediaItems))
	}
	expects := []string{
		"album1/a.jpg",
		"album2/b.jpg",
		"album2/c.jpg",
	}
	for i, expect := range expects {
		if mediaItems[i].String() != expect {
			t.Errorf("[%d] wants %s but %+v", i, expect, mediaItems[i])
		}
	}
	if mediaItem, ok := mediaItems[3].(*photos.HTTPMediaItem); !ok {
		t.Errorf("[3] wants HTTPMediaItem but %+v", mediaItems[3])
	} else if r := mediaItem.Request; r == nil {
		t.Errorf("[3].Request wants non-nil but nil")
	} else if want := "http://www.example.com/image.jpg"; r.URL.String() != want {
		t.Errorf("[3].Request.URL wants %s but %s", want, r.URL)
	}
}

func TestCLI_findMediaItems_Headers(t *testing.T) {
	c := CLI{
		Paths:            []string{"http://www.example.com/image.jpg"},
		RequestHeaders:   []string{"Cookie: foo"},
		RequestBasicAuth: "alice:bob",
	}
	mediaItems, err := c.findMediaItems()
	if err != nil {
		t.Fatal(err)
	}
	if len(mediaItems) != 1 {
		t.Errorf("wants size 1 but %d", len(mediaItems))
	}
	if mediaItem, ok := mediaItems[0].(*photos.HTTPMediaItem); !ok {
		t.Errorf("[0] wants HTTPMediaItem but %+v", mediaItems[0])
	} else if r := mediaItem.Request; r == nil {
		t.Errorf("[0].Request wants non-nil but nil")
	} else if r.URL.String() != "http://www.example.com/image.jpg" {
		t.Errorf("[0].Request.URL wants %s but %s", "http://www.example.com/image.jpg", r.URL)
	} else if v := r.Header.Get("Cookie"); v != "foo" {
		t.Errorf("[0].Header(Cookie) wants foo but %s", v)
	} else if username, password, ok := r.BasicAuth(); !ok {
		t.Errorf("[0].BasicAuth wants ok but not")
	} else if username != "alice" {
		t.Errorf("[0].BasicAuth.username wants alice but %s", username)
	} else if password != "bob" {
		t.Errorf("[0].BasicAuth.password wants bob but %s", password)
	}
}
