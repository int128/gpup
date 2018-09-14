package cli

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/int128/gpup/photos"
)

func (c *CLI) upload(ctx context.Context) error {
	if len(c.Paths) == 0 {
		return fmt.Errorf("Nothing to upload")
	}
	files, err := findFiles(c.Paths)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("File not found in %s", strings.Join(c.Paths, ", "))
	}
	log.Printf("The following %d files will be uploaded:", len(files))
	for i, file := range files {
		fmt.Printf("%3d: %s\n", i+1, file)
	}

	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}
	service, err := photos.New(client)
	if err != nil {
		return err
	}

	switch {
	case c.AlbumTitle != "":
		return service.AddToAlbum(ctx, c.AlbumTitle, files)
	case c.NewAlbum != "":
		_, err = service.CreateAlbum(ctx, c.NewAlbum, files)
		return err
	default:
		return service.AddToLibrary(ctx, files)
	}
}
