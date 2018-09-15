package cli

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/int128/gpup/photos"
)

func findMediaItems(args []string, client *http.Client) ([]photos.MediaItem, error) {
	mediaItems := make([]photos.MediaItem, 0)
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://"):
			r, err := http.NewRequest("GET", arg, nil)
			if err != nil {
				return nil, fmt.Errorf("Could not parse URL: %s", err)
			}
			mediaItems = append(mediaItems, &photos.HTTPMediaItem{
				Client:  client,
				Request: r,
			})
		default:
			if err := filepath.Walk(arg, func(name string, info os.FileInfo, err error) error {
				switch {
				case err != nil:
					return err
				case info.Mode().IsRegular():
					mediaItems = append(mediaItems, photos.FileMediaItem(name))
					return nil
				default:
					return nil
				}
			}); err != nil {
				return nil, fmt.Errorf("Error while finding files in %s: %s", arg, err)
			}
		}
	}
	return mediaItems, nil
}
