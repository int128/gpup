package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

func findFiles(filePaths []string) ([]string, error) {
	files := make([]string, 0, len(filePaths)*2)
	for _, parent := range filePaths {
		if err := filepath.Walk(parent, func(child string, info os.FileInfo, err error) error {
			switch {
			case err != nil:
				return err
			case info.Mode().IsRegular():
				files = append(files, child)
				return nil
			default:
				return nil
			}
		}); err != nil {
			return nil, fmt.Errorf("Error while finding files in %s: %s", parent, err)
		}
	}
	return files, nil
}
