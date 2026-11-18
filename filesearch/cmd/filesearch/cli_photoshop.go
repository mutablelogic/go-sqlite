//go:build !cgo || media

package main

import (
	// The handler of Photoshop documents, which is pure Go without cgo, but
	// which reads their EXIF with libexif with it, so it's built with cgo
	// only with the media tag, with the other C libraries
	_ "github.com/mutablelogic/go-metadata/image/photoshop"
)
