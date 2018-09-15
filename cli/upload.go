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
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}
	mediaItems, err := findMediaItems(c.Paths, client)
	if err != nil {
		return err
	}
	if len(mediaItems) == 0 {
		return fmt.Errorf("Nothing to upload in %s", strings.Join(c.Paths, ", "))
	}
	log.Printf("The following %d items will be uploaded:", len(mediaItems))
	for i, mediaItem := range mediaItems {
		fmt.Printf("%3d: %s\n", i+1, mediaItem)
	}

	service, err := photos.New(client)
	if err != nil {
		return err
	}
	switch {
	case c.AlbumTitle != "":
		return service.AddToAlbum(ctx, c.AlbumTitle, mediaItems)
	case c.NewAlbum != "":
		_, err = service.CreateAlbum(ctx, c.NewAlbum, mediaItems)
		return err
	default:
		return service.AddToLibrary(ctx, mediaItems)
	}
}
