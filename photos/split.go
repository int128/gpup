package photos

import (
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

func splitMediaItems(all []*photoslibrary.NewMediaItem, unit int) [][]*photoslibrary.NewMediaItem {
	chunkCount := len(all) / unit
	if len(all)%unit > 0 {
		chunkCount++
	}
	chunks := make([][]*photoslibrary.NewMediaItem, chunkCount)
	for i := 0; i < chunkCount; i++ {
		chunks[i] = all[i*unit : minInt((i+1)*unit, len(all))]
	}
	return chunks
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
