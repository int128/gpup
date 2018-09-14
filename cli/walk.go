package cli

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/int128/gpup/photos"
)

func findFiles(filePaths []string, client *http.Client) ([]photos.Media, error) {
	files := make([]photos.Media, 0)
	for _, parent := range filePaths {
		switch {
		case strings.HasPrefix(parent, "http://") || strings.HasPrefix(parent, "https://"):
			r, err := http.NewRequest("GET", parent, nil)
			if err != nil {
				return nil, fmt.Errorf("Could not parse URL: %s", err)
			}
			media := photos.HTTPMedia{
				Client:  client,
				Request: r,
			}
			files = append(files, &media)
		default:
			if err := filepath.Walk(parent, func(child string, info os.FileInfo, err error) error {
				switch {
				case err != nil:
					return err
				case info.Mode().IsRegular():
					files = append(files, photos.FileMedia(child))
					return nil
				default:
					return nil
				}
			}); err != nil {
				return nil, fmt.Errorf("Error while finding files in %s: %s", parent, err)
			}
		}
	}
	return files, nil
}
