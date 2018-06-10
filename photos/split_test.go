package photos

import (
	"testing"

	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

func TestSplitMediaItems(t *testing.T) {
	matrix := []struct {
		length   int
		unit     int
		expected []int
	}{
		{0, 1, []int{}},
		{1, 1, []int{1}},
		{2, 1, []int{1, 1}},
		{0, 2, []int{}},
		{1, 2, []int{1}},
		{2, 2, []int{2}},
		{3, 2, []int{2, 1}},
		{0, 10, []int{}},
		{1, 10, []int{1}},
		{9, 10, []int{9}},
		{10, 10, []int{10}},
		{11, 10, []int{10, 1}},
		{49, 10, []int{10, 10, 10, 10, 9}},
		{50, 10, []int{10, 10, 10, 10, 10}},
		{51, 10, []int{10, 10, 10, 10, 10, 1}},
	}
	for _, m := range matrix {
		items := make([]*photoslibrary.NewMediaItem, m.length)
		actual := splitMediaItems(items, m.unit)
		if len(actual) != len(m.expected) {
			t.Errorf("chunk count should be %d but %d", len(m.expected), len(actual))
		}
		for i, actualChunk := range actual {
			if len(actualChunk) != m.expected[i] {
				t.Errorf("length of chunk[%d] should be %d but %d", i, m.expected[i], len(actualChunk))
			}
		}
	}
}
