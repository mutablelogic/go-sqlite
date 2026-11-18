package filesearch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	// Packages
	metadata "github.com/mutablelogic/go-metadata"
	types "github.com/mutablelogic/go-server/pkg/types"
	sqlite "github.com/mutablelogic/go-sqlite"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
	test "github.com/mutablelogic/go-sqlite/filesearch/test"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"

	// The handlers of go-metadata, which are pure Go, which read text, such
	// as Markdown, and images, such as PNG
	_ "github.com/mutablelogic/go-metadata/image/raster"
	_ "github.com/mutablelogic/go-metadata/text"
)

// pngImage returns a PNG image, of a size, in one colour
func pngImage(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func Test_ReadFiles(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()
	write(t, root, map[string]string{
		"notes/rain.md": "# Replicants in the Rain\n\nAll those moments will be lost in time, like tears in rain.\n",
		"plain.txt":     "Nothing to see here, but quietly.",
	})
	require.NoError(t, os.WriteFile(filepath.Join(root, "photo.png"), pngImage(t, 1024, 512), 0o644))

	// They're read, of their metadata, content and artwork, with the
	// handlers which are imported
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root, Recursive: true}, nil)
	require.NoError(t, err)
	list, err := manager.ListObjects(ctx, schema.ObjectListRequest{Volumes: []int64{volume.ID}})
	require.NoError(t, err)
	byPath := make(map[string]schema.Object)
	for _, object := range list.Body {
		byPath[object.Path] = object
	}

	// A Markdown document, of its heading, and its content, which is searched
	note := byPath["notes/rain.md"]
	assert.Equal(t, "text/markdown", note.Type)
	assert.Equal(t, "Replicants in the Rain", note.Title)
	assert.NotContains(t, note.Meta, "metadata:hint", "what to read next isn't kept")
	for query, want := range map[string]string{"tears": "notes/rain.md", "quiet": "plain.txt", "replicants": "notes/rain.md"} {
		found, err := manager.ListObjects(ctx, schema.ObjectListRequest{Query: types.Ptr(query)})
		require.NoError(t, err)
		if assert.Len(t, found.Body, 1, query) {
			assert.Equal(t, want, found.Body[0].Path, query)
		}
	}

	// An image, of its kind and dimensions, and its artwork, which is
	// scaled down
	photo := byPath["photo.png"]
	assert.Equal(t, "image/png", photo.Type)
	assert.NotEmpty(t, photo.Kind)
	assert.NotEmpty(t, photo.Meta)
	artwork, err := manager.Artwork(ctx, photo.ID)
	require.NoError(t, err)
	img, _, err := image.Decode(bytes.NewReader(artwork))
	require.NoError(t, err)
	assert.LessOrEqual(t, max(img.Bounds().Dx(), img.Bounds().Dy()), 256)
	assert.Equal(t, 2, img.Bounds().Dx()/img.Bounds().Dy(), "of its aspect ratio")
}

func Test_ReadFileUnreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permissions aren't enforced")
	}
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()
	write(t, root, map[string]string{"a.txt": "a", "locked.txt": "b"})
	locked := filepath.Join(root, "locked.txt")
	require.NoError(t, os.Chmod(locked, 0o000))
	t.Cleanup(func() { _ = os.Chmod(locked, 0o644) })

	// A file which can't be opened is said, and isn't indexed
	var errs []error
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, func(p schema.Progress, err error) {
		if err != nil {
			errs = append(errs, err)
		}
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"a.txt"}, objects(t, manager, volume.ID))
	if assert.Len(t, errs, 1) {
		assert.ErrorIs(t, errs[0], os.ErrPermission)
	}
}

func Test_ReadFileErrors(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()

	// A handler, of a type of file, which has content, as Markdown, and fails
	// to read it, which is its error, of its metadata, and it's indexed
	require.NoError(t, metadata.AddTextExtensionType(".fstest", "text/x-filesearch-test"))
	metadata.AddHandler(regexp.MustCompile(`^text/x-filesearch-test$`), "filesearch-test", func(ctx context.Context, r io.Reader, opts *metadata.Opts) ([]metadata.Metadata, error) {
		if opts.HasNamespace(metadata.MarkdownNamespace) {
			return nil, errors.New("its markdown failed")
		}
		return metadata.HintMetadata(metadata.HintMarkdown), nil
	}, metadata.MetadataNamespace, metadata.MarkdownNamespace)
	write(t, root, map[string]string{"broken.fstest": "content"})
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, func(p schema.Progress, err error) {
		assert.NoError(t, err, "a handler's error isn't the run's")
	})
	require.NoError(t, err)
	list, err := manager.ListObjects(ctx, schema.ObjectListRequest{Volumes: []int64{volume.ID}})
	require.NoError(t, err)
	if assert.Len(t, list.Body, 1) {
		assert.Contains(t, list.Body[0].Meta["filesearch:error"], "its markdown failed")
	}
}

func Test_GetObject(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)
	a, b := t.TempDir(), t.TempDir()
	write(t, a, map[string]string{"notes/rain.md": "# Rain\n\nIt rained.\n", "notes/sun.md": "# Sun\n"})
	write(t, b, map[string]string{"rain.md": "# Other Rain\n"})
	va, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: a, Recursive: true}, nil)
	require.NoError(t, err)
	_, err = manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: b}, nil)
	require.NoError(t, err)

	// A file, of its full path, its path within its volume, or its name,
	// when it's of one file
	for _, path := range []string{filepath.Join(va.Path, "notes", "sun.md"), "notes/sun.md", "sun.md", "notes/sun.md/"} {
		object, err := manager.GetObject(ctx, path)
		require.NoError(t, err, path)
		assert.Equal(t, "notes/sun.md", object.Path, path)
		full, err := manager.ObjectPath(ctx, object)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(va.Path, "notes", "sun.md"), full)
	}

	// Of more than one, it's a conflict, of their paths, and of none, not
	// found
	_, err = manager.GetObject(ctx, "rain.md")
	assert.ErrorIs(t, err, sqlite.ErrConflict)
	assert.ErrorContains(t, err, "2 files")
	assert.ErrorContains(t, err, filepath.Join(va.Path, "notes", "rain.md"))
	object, err := manager.GetObject(ctx, "notes/rain.md")
	require.NoError(t, err, "of one, within its volume")
	_, err = manager.GetObject(ctx, "missing.md")
	assert.ErrorIs(t, err, sqlite.ErrNotFound)
	_, err = manager.GetObject(ctx, "")
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)

	// Its content, as Markdown, which is read again, all of it
	markdown, err := manager.Markdown(ctx, object.ID)
	require.NoError(t, err)
	assert.Contains(t, markdown, "It rained.")
}

// durationEntry is metadata of a duration, which marshals it as text, as
// go-metadata's entries do
type durationEntry time.Duration

func (e durationEntry) NS() string                   { return "fstest" }
func (e durationEntry) Key() string                  { return "runtime" }
func (e durationEntry) Value() string                { return time.Duration(e).String() }
func (e durationEntry) Bytes() []byte                { return []byte(e.Value()) }
func (e durationEntry) Image() image.Image           { return nil }
func (e durationEntry) Any() any                     { return time.Duration(e) }
func (e durationEntry) MarshalJSON() ([]byte, error) { return json.Marshal(e.Value()) }

// nameEntry is metadata of a name, which marshals itself, as go-metadata's
// entries do
type nameEntry string

func (e nameEntry) NS() string                   { return "fstest" }
func (e nameEntry) Key() string                  { return "director" }
func (e nameEntry) Value() string                { return string(e) }
func (e nameEntry) Bytes() []byte                { return []byte(e) }
func (e nameEntry) Image() image.Image           { return nil }
func (e nameEntry) Any() any                     { return string(e) }
func (e nameEntry) MarshalJSON() ([]byte, error) { return json.Marshal(string(e)) }

// countEntry is metadata of a number, which doesn't marshal itself
type countEntry int

func (e countEntry) NS() string         { return "fstest" }
func (e countEntry) Key() string        { return "count" }
func (e countEntry) Value() string      { return fmt.Sprint(int(e)) }
func (e countEntry) Bytes() []byte      { return []byte(e.Value()) }
func (e countEntry) Image() image.Image { return nil }
func (e countEntry) Any() any           { return int(e) }

func Test_ReadFileValues(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()

	// A value is as its entry marshals it, such as a duration as text, or,
	// when it doesn't, as it is
	require.NoError(t, metadata.AddTextExtensionType(".fsvalues", "text/x-filesearch-values"))
	metadata.AddHandler(regexp.MustCompile(`^text/x-filesearch-values$`), "filesearch-values", func(ctx context.Context, r io.Reader, opts *metadata.Opts) ([]metadata.Metadata, error) {
		return []metadata.Metadata{durationEntry(118 * time.Minute), countEntry(3), nameEntry("Ridley Scott")}, nil
	}, metadata.MetadataNamespace)
	write(t, root, map[string]string{"film.fsvalues": "content"})
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, nil)
	require.NoError(t, err)
	list, err := manager.ListObjects(ctx, schema.ObjectListRequest{Volumes: []int64{volume.ID}})
	require.NoError(t, err)
	if assert.Len(t, list.Body, 1) {
		assert.Equal(t, "1h58m0s", list.Body[0].Meta["fstest:runtime"])
		assert.EqualValues(t, 3, list.Body[0].Meta["fstest:count"])
		assert.Equal(t, "Ridley Scott", list.Body[0].Meta["fstest:director"])
	}

	// A string which marshals itself is searched, though it's only of the
	// metadata
	found, err := manager.ListObjects(ctx, schema.ObjectListRequest{Query: types.Ptr("ridley"), Cutoff: types.Ptr(0.0)})
	require.NoError(t, err)
	if assert.Len(t, found.Body, 1) {
		assert.Equal(t, "film.fsvalues", found.Body[0].Path)
	}
}
