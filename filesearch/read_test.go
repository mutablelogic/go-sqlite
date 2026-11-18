package filesearch_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
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
