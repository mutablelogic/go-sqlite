//go:build media

package main

import (
	// The handlers of go-metadata, of audio and video, and photos' EXIF, HEIF
	// and RAW, which need cgo, and its C libraries: FFmpeg 8, libexif,
	// libheif and LibRaw (see go-metadata's README). Build with:
	//
	//   go build -tags media ./cmd/filesearch
	_ "github.com/mutablelogic/go-metadata/audio"
	_ "github.com/mutablelogic/go-metadata/image"
	_ "github.com/mutablelogic/go-metadata/video"
)
