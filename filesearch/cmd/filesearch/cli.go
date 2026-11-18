package main

import (
	// Packages
	filesearchcmd "github.com/mutablelogic/go-sqlite/filesearch/cmd"

	// The handlers of go-metadata, which read files, of their types, which
	// are pure Go: documents, email, text, images, such as PNG, JPEG and
	// SVG, playlists, subtitles, the kinds of audio and video, and the movies
	// and episodes on TMDB which videos are, with --tmdb-key. Photoshop's is
	// too, without cgo (see cli_photoshop.go). The others, of audio and
	// video, and photos' EXIF, HEIF and RAW, are built with the media tag
	// (see cli_media.go).
	_ "github.com/mutablelogic/go-metadata/application"
	_ "github.com/mutablelogic/go-metadata/audio/kind"
	_ "github.com/mutablelogic/go-metadata/audio/playlist"
	_ "github.com/mutablelogic/go-metadata/image/raster"
	_ "github.com/mutablelogic/go-metadata/image/svg"
	_ "github.com/mutablelogic/go-metadata/message"
	_ "github.com/mutablelogic/go-metadata/text"
	_ "github.com/mutablelogic/go-metadata/video/kind"
	_ "github.com/mutablelogic/go-metadata/video/subtitle"
	_ "github.com/mutablelogic/go-metadata/video/tmdb"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// CLI is the commands of filesearch
type CLI struct {
	filesearchcmd.Commands
}
