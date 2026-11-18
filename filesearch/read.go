package filesearch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	// Packages
	metadata "github.com/mutablelogic/go-metadata"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// read is what's read of a file: its content type, kind and title, its
// metadata, of "ns:key", its content, as Markdown, and its artwork
type read struct {
	contentType string
	kind        string
	title       string
	meta        map[string]any
	text        string
	artwork     []byte
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// MaxText is the most of a file's content, as Markdown, which is read,
	// and searched, in bytes
	MaxText = 256 * 1024

	// ArtworkSize is the largest width or height of a file's artwork, in
	// pixels
	ArtworkSize = 256

	// The metadata which isn't a file's own: what to read of it next
	hintKey = metadata.MetadataNamespace + ":" + metadata.HintKey
	kindKey = metadata.MetadataNamespace + ":" + metadata.KindKey

	// The metadata of an error, which go-metadata returned, as it read a file
	errorKey = "filesearch:error"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// readFile reads a file, with go-metadata, of the handlers which are
// imported: its metadata, and then, when its hints are to, its content, as
// Markdown, and its artwork, so a file which has neither isn't read twice.
// A file which no handler reads has its content type, and its title, of its
// name, and a handler's error is in its metadata, as metadata:note, so
// neither is an error. Only a file which can't be opened, or a context
// which is done, is.
func (m *Manager) readFile(ctx context.Context, path string) (*read, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Its metadata, of the namespaces which aren't requested on their own
	contentType, entries, err := metadata.Extract(ctx, f, metadata.WithTracer(m.tracer))
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := &read{contentType: contentType, meta: make(map[string]any)}
	var hints []string
	for _, entry := range entries {
		key := entry.NS() + ":" + entry.Key()
		switch {
		case key == hintKey:
			hints = stringList(entry.Any())
		case key == kindKey:
			result.kind = entry.Value()
		case entry.NS() == "dc" && entry.Key() == "title" && result.title == "":
			result.title = entry.Value()
		}
		if key != hintKey {
			addMeta(result.meta, key, entry)
		}
	}
	if err != nil && !errors.Is(err, metadata.ErrNotImplemented) {
		result.meta[errorKey] = err.Error()
	}

	// Its content, as Markdown, and its artwork, when it has them
	var namespaces []string
	if slices.Contains(hints, metadata.HintMarkdown) {
		namespaces = append(namespaces, metadata.MarkdownNamespace)
	}
	if slices.Contains(hints, metadata.HintArtwork) {
		namespaces = append(namespaces, metadata.ArtworkNamespace)
	}
	if len(namespaces) == 0 {
		return result, nil
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	_, entries, _ = metadata.Extract(ctx, f,
		metadata.WithTracer(m.tracer),
		metadata.WithContentType(contentType),
		metadata.WithNamespace(namespaces...),
		metadata.WithRange(fmt.Sprintf("0:%d", MaxText)),
		metadata.WithMaxImageDimension(ArtworkSize),
	)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var text strings.Builder
	for _, entry := range entries {
		switch {
		case entry.NS() == metadata.MarkdownNamespace && entry.Key() == "body":
			text.WriteString(entry.Value())
		case entry.NS() == metadata.ArtworkNamespace && result.artwork == nil && entry.Image() != nil:
			result.artwork = entry.Bytes()
		}
	}
	result.text = text.String()
	return result, nil
}

// addMeta adds a value of metadata, of its key, which is a list of the
// values, when there's more than one, unless it's an image, or bytes, which
// aren't searched
func addMeta(meta map[string]any, key string, entry metadata.Metadata) {
	if entry.Image() != nil {
		return
	}
	v := entry.Any()
	if _, isBytes := v.([]byte); isBytes || v == nil {
		return
	}
	switch existing := meta[key].(type) {
	case nil:
		meta[key] = v
	case []any:
		meta[key] = append(existing, v)
	default:
		meta[key] = []any{existing, v}
	}
}

// stringList returns a value which is a list of strings, or a string, as a
// list
func stringList(v any) []string {
	switch v := v.(type) {
	case []string:
		return v
	case string:
		return []string{v}
	case []any:
		var result []string
		for _, v := range v {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}
