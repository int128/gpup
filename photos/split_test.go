package photos

import (
	"fmt"
	"strings"
	"testing"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

var alphabets = strings.Split("abcdefghijklmnopqr", "")

type chunk []string
type chunks []chunk

func TestSplitMediaItems(t *testing.T) {
	matrix := []struct {
		length         int
		unit           int
		expectedChunks chunks
	}{
		{0, 1, chunks{}},
		{0, 2, chunks{}},
		{0, 3, chunks{}},
		{1, 1, chunks{{"a"}}},
		{1, 2, chunks{{"a"}}},
		{1, 3, chunks{{"a"}}},
		{2, 1, chunks{{"a"}, {"b"}}},
		{2, 2, chunks{{"a", "b"}}},
		{2, 3, chunks{{"a", "b"}}},
		{3, 1, chunks{{"a"}, {"b"}, {"c"}}},
		{3, 2, chunks{{"a", "b"}, {"c"}}},
		{3, 3, chunks{{"a", "b", "c"}}},
		{3, 4, chunks{{"a", "b", "c"}}},
		{3, 100, chunks{{"a", "b", "c"}}},
		{4, 1, chunks{{"a"}, {"b"}, {"c"}, {"d"}}},
		{4, 2, chunks{{"a", "b"}, {"c", "d"}}},
		{4, 3, chunks{{"a", "b", "c"}, {"d"}}},
		{4, 4, chunks{{"a", "b", "c", "d"}}},
		{4, 5, chunks{{"a", "b", "c", "d"}}},
		{4, 50, chunks{{"a", "b", "c", "d"}}},
		{5, 1, chunks{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}}},
	}
	for _, m := range matrix {
		t.Run(fmt.Sprintf("length=%d/unit=%d", m.length, m.unit), func(t *testing.T) {
			items := createItems(t, m.length)
			actualChunks := splitMediaItems(items, m.unit)
			if len(m.expectedChunks) != len(actualChunks) {
				t.Errorf("chunks count wants %d but %d", len(m.expectedChunks), len(actualChunks))
			}
			for i := range actualChunks {
				if len(m.expectedChunks[i]) != len(actualChunks[i]) {
					t.Errorf("length of chunk[%d] wants %d but %d", i, len(m.expectedChunks[i]), len(actualChunks[i]))
				}
				for j := range actualChunks[i] {
					if m.expectedChunks[i][j] != actualChunks[i][j].Description {
						t.Errorf("chunk[%d][%d] wants %s but %s", i, j, m.expectedChunks[i][j], actualChunks[i][j].Description)
					}
				}
			}
		})
	}
}

func createItems(t *testing.T, length int) []*photoslibrary.NewMediaItem {
	t.Helper()
	items := make([]*photoslibrary.NewMediaItem, length)
	for i := range items {
		items[i] = &photoslibrary.NewMediaItem{Description: alphabets[i]}
	}
	return items
}
