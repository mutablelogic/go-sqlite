package filesearch_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	filesearch "github.com/mutablelogic/go-sqlite/filesearch"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
	test "github.com/mutablelogic/go-sqlite/filesearch/test"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// write writes files, of their paths, within the folder, with their
// contents
func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		path = filepath.Join(root, filepath.FromSlash(path))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
}

// objects returns the paths of the objects of a volume, in order
func objects(t *testing.T, manager *filesearch.Manager, volume int64) []string {
	t.Helper()
	list, err := manager.ListObjects(t.Context(), schema.ObjectListRequest{Volume: volume})
	require.NoError(t, err)
	var paths []string
	for _, o := range list.Body {
		paths = append(paths, o.Path)
	}
	return paths
}

func Test_IndexVolume(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()
	write(t, root, map[string]string{
		"a.txt":         "a",
		"notes/b.md":    "# B",
		"photos/c.jpeg": "c",
	})

	// A volume is indexed, of its files and folders, as it's recursive, with
	// their types and titles, and its progress is said
	var said []schema.Progress
	volume, err := manager.IndexVolume(ctx, schema.VolumeMeta{Path: root, Recursive: true}, func(p schema.Progress, err error) {
		assert.NoError(t, err)
		said = append(said, p)
	})
	require.NoError(t, err)
	assert.Equal(t, schema.StatusIndexed, volume.Status)
	assert.Equal(t, int64(1), volume.Run)
	assert.Equal(t, []string{"a.txt", "notes/b.md", "photos/c.jpeg"}, objects(t, manager, volume.ID))
	if assert.Len(t, said, 4, "three files, and finished") {
		assert.Equal(t, uint64(3), said[0].Files, "the files of a batch are walked, and then indexed")
		assert.Equal(t, uint64(1), said[0].Indexed)
		assert.Nil(t, said[0].Finished)
		assert.NotNil(t, said[3].Finished)
		assert.Equal(t, uint64(3), said[3].Indexed)
	}
	list, err := manager.ListObjects(ctx, schema.ObjectListRequest{Query: "c"})
	require.NoError(t, err)
	if assert.Len(t, list.Body, 1) {
		assert.Equal(t, "image/jpeg", list.Body[0].Type)
		assert.Equal(t, "c", list.Body[0].Title)
	}

	// Its progress, once it's finished, is how it did
	progress, err := manager.Progress(volume.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), progress.Files)
	assert.Equal(t, uint64(3), progress.Indexed)
	assert.Zero(t, progress.Marked)
	assert.NotNil(t, progress.Finished)
	assert.Empty(t, progress.Current)
	_, err = manager.Progress(volume.ID + 1)
	assert.ErrorIs(t, err, sqlite.ErrNotFound)

	// Indexed again, what's as it was is marked, what's changed, or new, is
	// indexed, and what's gone is swept
	write(t, root, map[string]string{"notes/b.md": "# B, changed", "d.txt": "d"})
	require.NoError(t, os.Remove(filepath.Join(root, "a.txt")))
	again, err := manager.IndexVolume(ctx, schema.VolumeMeta{Path: root, Recursive: true}, nil)
	require.NoError(t, err)
	assert.Equal(t, volume.ID, again.ID)
	assert.Equal(t, int64(2), again.Run)
	assert.Equal(t, []string{"d.txt", "notes/b.md", "photos/c.jpeg"}, objects(t, manager, volume.ID))
	progress, err = manager.Progress(volume.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), progress.Files)
	assert.Equal(t, uint64(2), progress.Indexed)
	assert.Equal(t, uint64(1), progress.Marked)
	assert.Equal(t, uint64(1), progress.Swept)

	// Indexed again, not recursively, its folders' files are swept
	again, err = manager.IndexVolume(ctx, schema.VolumeMeta{Path: root}, nil)
	require.NoError(t, err)
	assert.False(t, again.Recursive)
	assert.Equal(t, []string{"d.txt"}, objects(t, manager, volume.ID))
}

func Test_IndexVolumeErrors(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()
	write(t, root, map[string]string{"a.txt": "a", "b.txt": "b"})

	// A path which isn't a folder isn't a volume
	_, err := manager.IndexVolume(ctx, schema.VolumeMeta{Path: filepath.Join(root, "a.txt")}, nil)
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = manager.IndexVolume(ctx, schema.VolumeMeta{Path: filepath.Join(root, "missing")}, nil)
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)

	// A volume which is indexing is a conflict, and its progress is where it
	// is, until it's finished
	var once sync.Once
	var indexing int64
	walking, finish := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	var volume *schema.Volume
	go func() {
		var err error
		volume, err = manager.IndexVolume(ctx, schema.VolumeMeta{Path: root}, func(p schema.Progress, err error) {
			once.Do(func() {
				indexing = p.Volume
				close(walking)
				<-finish
			})
		})
		done <- err
	}()
	<-walking
	_, err = manager.IndexVolume(ctx, schema.VolumeMeta{Path: root}, nil)
	assert.ErrorIs(t, err, sqlite.ErrConflict)
	progress, err := manager.Progress(indexing)
	require.NoError(t, err)
	assert.Nil(t, progress.Finished)
	assert.Equal(t, root, progress.Path)
	assert.NotEmpty(t, progress.Current)
	assert.Equal(t, uint64(1), progress.Indexed)

	// What it's indexing isn't seen, until its batch is committed
	list, err := manager.ListObjects(ctx, schema.ObjectListRequest{})
	require.NoError(t, err)
	assert.Empty(t, list.Body)
	close(finish)
	require.NoError(t, <-done)
	assert.Equal(t, schema.StatusIndexed, volume.Status)

	// A run whose context is done fails, and nothing is swept
	write(t, root, map[string]string{"c.txt": "c"})
	cancelled, cancel := context.WithCancel(ctx)
	var errs []error
	failed, err := manager.IndexVolume(cancelled, schema.VolumeMeta{Path: root}, func(p schema.Progress, err error) {
		cancel()
		errs = append(errs, err)
	})
	assert.ErrorIs(t, err, context.Canceled)
	require.NotNil(t, failed)
	assert.Equal(t, schema.StatusFailed, failed.Status)
	assert.Contains(t, failed.Error, "canceled")
	assert.ErrorIs(t, errs[len(errs)-1], context.Canceled, "it's said, as it finished")
	assert.Equal(t, []string{"a.txt", "b.txt"}, objects(t, manager, volume.ID), "nothing is swept, or indexed, after it's cancelled")
}

func Test_IndexVolumeUnreadable(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permissions aren't enforced")
	}
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()
	write(t, root, map[string]string{"a.txt": "a", "locked/b.txt": "b"})
	locked := filepath.Join(root, "locked")
	require.NoError(t, os.Chmod(locked, 0o000))
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

	// A folder which can't be read is an error, which is said, and the others
	// are indexed
	var errs []error
	volume, err := manager.IndexVolume(ctx, schema.VolumeMeta{Path: root, Recursive: true}, func(p schema.Progress, err error) {
		if err != nil {
			errs = append(errs, err)
		}
	})
	require.NoError(t, err)
	assert.Equal(t, schema.StatusIndexed, volume.Status)
	assert.Equal(t, []string{"a.txt"}, objects(t, manager, volume.ID))
	if assert.Len(t, errs, 1) {
		assert.ErrorIs(t, errs[0], os.ErrPermission)
		assert.ErrorContains(t, errs[0], "locked")
	}
}

func Test_IndexVolumeBatches(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()

	// More files than a batch, which are all indexed, and then marked
	files := make(map[string]string)
	for i := range 250 {
		files[filepath.Join("files", time.Duration(i).String()+".txt")] = "x"
	}
	write(t, root, files)
	volume, err := manager.IndexVolume(ctx, schema.VolumeMeta{Path: root, Recursive: true}, nil)
	require.NoError(t, err)
	paths := objects(t, manager, volume.ID)
	assert.Len(t, paths, 100, "a page of them")
	list, err := manager.ListObjects(ctx, schema.ObjectListRequest{Volume: volume.ID})
	require.NoError(t, err)
	assert.Equal(t, uint64(250), list.Count)
	_, err = manager.IndexVolume(ctx, schema.VolumeMeta{Path: root, Recursive: true}, nil)
	require.NoError(t, err)
	progress, err := manager.Progress(volume.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(250), progress.Marked)
	assert.True(t, sort.StringsAreSorted(paths))
}
