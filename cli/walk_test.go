package cli

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestFindFiles(t *testing.T) {
	expects := []string{
		"album1/a.jpg",
		"album2/b.jpg",
		"album2/c.jpg",
	}

	tempdir, err := ioutil.TempDir("", "FindFiles")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)
	if err := os.Mkdir(tempdir+"/album1", 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(tempdir+"/album1/a.jpg", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(tempdir+"/album2", 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(tempdir+"/album2/b.jpg", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(tempdir+"/album2/c.jpg", []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	os.Chdir(tempdir)
	if err != nil {
		t.Fatal(err)
	}

	files, err := findMediaItems([]string{"."}, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	for i, expect := range expects {
		if files[i].String() != expect {
			t.Errorf("files[%d] wants %s but %s", i, expect, files[0])
		}
	}
}
