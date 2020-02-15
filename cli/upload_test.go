package cli

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/cheggaaa/pb/v3"
	"github.com/int128/gpup/photos"
)

func TestCLI_findUploadItems(t *testing.T) {
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
	bar := pb.New(0)
	uploadItems, err := c.findUploadItems(bar)
	if err != nil {
		t.Fatal(err)
	}
	if len(uploadItems) != 4 {
		t.Errorf("wants size 4 but %d", len(uploadItems))
	}
	expects := []string{
		"album1/a.jpg",
		"album2/b.jpg",
		"album2/c.jpg",
	}
	for i, expect := range expects {
		if uploadItems[i].String() != expect {
			t.Errorf("[%d] wants %s but %+v", i, expect, uploadItems[i])
		}
	}
	if uploadItem, ok := uploadItems[3].(*photos.HTTPUploadItem); !ok {
		t.Errorf("[3] wants HTTPUploadItem but %+v", uploadItems[3])
	} else if r := uploadItem.Request; r == nil {
		t.Errorf("[3].Request wants non-nil but nil")
	} else if want := "http://www.example.com/image.jpg"; r.URL.String() != want {
		t.Errorf("[3].Request.URL wants %s but %s", want, r.URL)
	}
}

func TestCLI_findUploadItems_Headers(t *testing.T) {
	c := CLI{
		Paths:            []string{"http://www.example.com/image.jpg"},
		RequestHeaders:   []string{"Cookie: foo"},
		RequestBasicAuth: "alice:bob",
	}
	bar := pb.New(0)
	uploadItems, err := c.findUploadItems(bar)
	if err != nil {
		t.Fatal(err)
	}
	if len(uploadItems) != 1 {
		t.Errorf("wants size 1 but %d", len(uploadItems))
	}
	if uploadItem, ok := uploadItems[0].(*photos.HTTPUploadItem); !ok {
		t.Errorf("[0] wants HTTPUploadItem but %+v", uploadItems[0])
	} else if r := uploadItem.Request; r == nil {
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
