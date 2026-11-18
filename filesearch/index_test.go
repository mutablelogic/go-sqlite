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
	types "github.com/mutablelogic/go-server/pkg/types"
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
	list, err := manager.ListObjects(t.Context(), schema.ObjectListRequest{Volumes: []int64{volume}})
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
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root, Recursive: true}, func(p schema.Progress, err error) {
		assert.NoError(t, err)
		said = append(said, p)
	})
	require.NoError(t, err)
	assert.Equal(t, schema.StatusIndexed, volume.Status)
	assert.Equal(t, int64(1), volume.Run)
	assert.Equal(t, []string{"a.txt", "notes/b.md", "photos/c.jpeg"}, objects(t, manager, volume.ID))
	if assert.Len(t, said, 5, "counted, three files, and finished") {
		assert.Equal(t, uint64(3), said[0].Total, "counted, before it's walked")
		assert.Zero(t, said[0].Files)
		assert.Equal(t, uint64(3), said[1].Files, "the files of a batch are walked, and then indexed")
		assert.Equal(t, uint64(1), said[1].Indexed)
		assert.Nil(t, said[1].Finished)
		assert.NotNil(t, said[4].Finished)
		assert.Equal(t, uint64(3), said[4].Indexed)
	}
	list, err := manager.ListObjects(ctx, schema.ObjectListRequest{Query: types.Ptr("c")})
	require.NoError(t, err)
	if assert.Len(t, list.Body, 1) {
		assert.Equal(t, "text/plain", list.Body[0].Type, "of its content, without its charset")
		assert.Equal(t, "c", list.Body[0].Title)
	}

	// A list has its request, with its limit clamped to what's returned,
	// and the caller's request isn't changed
	list, err = manager.ListObjects(ctx, schema.ObjectListRequest{Volumes: []int64{volume.ID}})
	require.NoError(t, err)
	assert.Equal(t, []int64{volume.ID}, list.Volumes)
	if assert.NotNil(t, list.Limit) {
		assert.Equal(t, uint64(3), *list.Limit, "of the default, clamped to the count")
	}
	limit := uint64(2)
	req := schema.ObjectListRequest{Volumes: []int64{volume.ID}, OffsetLimit: sqlite.OffsetLimit{Offset: 2, Limit: &limit}}
	list, err = manager.ListObjects(ctx, req)
	require.NoError(t, err)
	assert.Len(t, list.Body, 1)
	assert.Equal(t, uint64(2), list.Offset)
	assert.Equal(t, uint64(1), *list.Limit, "of the one which is returned")
	assert.Equal(t, uint64(2), limit, "the caller's")
	list, err = manager.ListObjects(ctx, schema.ObjectListRequest{Query: types.Ptr("c"), OffsetLimit: sqlite.OffsetLimit{Offset: 5}})
	require.NoError(t, err)
	assert.Empty(t, list.Body)
	assert.Equal(t, "c", *list.Query)
	assert.Equal(t, uint64(0), *list.Limit, "past the end")

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
	again, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root, Recursive: true}, nil)
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
	_, err = manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"d.txt"}, objects(t, manager, volume.ID))
}

func Test_IndexVolumeErrors(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)
	root := resolved(t, t.TempDir())
	write(t, root, map[string]string{"a.txt": "a", "b.txt": "b"})

	// A path which isn't a folder isn't a volume
	_, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: filepath.Join(root, "a.txt")}, nil)
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: filepath.Join(root, "missing")}, nil)
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
		volume, err = manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, func(p schema.Progress, err error) {
			if p.Indexed+p.Marked == 0 {
				return
			}
			once.Do(func() {
				indexing = p.Volume
				close(walking)
				<-finish
			})
		})
		done <- err
	}()
	<-walking
	_, err = manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, nil)
	assert.ErrorIs(t, err, sqlite.ErrConflict)
	progress, err := manager.Progress(indexing)
	require.NoError(t, err)
	assert.Nil(t, progress.Finished)
	assert.Equal(t, root, progress.Path)
	assert.NotEmpty(t, progress.Current)
	assert.Equal(t, uint64(1), progress.Indexed)

	// The function is called once its batch is committed, so what it's said
	// is seen
	list, err := manager.ListObjects(ctx, schema.ObjectListRequest{})
	require.NoError(t, err)
	assert.Len(t, list.Body, 2)
	close(finish)
	require.NoError(t, <-done)
	assert.Equal(t, schema.StatusIndexed, volume.Status)

	// A run whose context is done, before it's finished, fails, with its
	// error, and nothing is swept, and the next run sweeps it
	write(t, root, map[string]string{"c.txt": "c"})
	require.NoError(t, os.Remove(filepath.Join(root, "b.txt")))
	cancelled, cancel := context.WithCancel(ctx)
	var errs []error
	failed, err := manager.IndexVolume(cancelled, schema.IndexVolumeRequest{Path: root}, func(p schema.Progress, err error) {
		cancel()
		errs = append(errs, err)
	})
	assert.ErrorIs(t, err, context.Canceled)
	require.NotNil(t, failed)
	assert.Equal(t, schema.StatusFailed, failed.Status)
	assert.Contains(t, failed.Error, "canceled")
	assert.ErrorIs(t, errs[len(errs)-1], context.Canceled, "it's said, as it finished")
	assert.Equal(t, []string{"a.txt", "b.txt"}, objects(t, manager, volume.ID), "what's changed isn't read, once it's cancelled, and nothing is swept")
	again, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, nil)
	require.NoError(t, err)
	assert.Equal(t, schema.StatusIndexed, again.Status)
	assert.Empty(t, again.Error)
	assert.Equal(t, []string{"a.txt", "c.txt"}, objects(t, manager, volume.ID))
}

func Test_IndexVolumeProgress(t *testing.T) {
	ctx := t.Context()
	root := t.TempDir()
	write(t, root, map[string]string{"a.txt": "a", "b.txt": "b", "c.txt": "c"})

	// Of a database in memory, which is one connection, the function can
	// list what's indexed, as it's called once its batch is committed, and
	// it's said of each file
	manager, err := filesearch.New(ctx, sqlite.Memory)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, manager.Close()) })
	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var current []string
	volume, err := manager.IndexVolume(timeout, schema.IndexVolumeRequest{Path: root}, func(p schema.Progress, err error) {
		require.NoError(t, err)
		if p.Finished != nil || p.Current == "" {
			return
		}
		current = append(current, p.Current)
		list, err := manager.ListObjects(timeout, schema.ObjectListRequest{})
		require.NoError(t, err)
		assert.Len(t, list.Body, 3)
	})
	require.NoError(t, err)
	assert.Equal(t, schema.StatusIndexed, volume.Status)
	assert.Equal(t, []string{"a.txt", "b.txt", "c.txt"}, current)
}

func Test_IndexVolumeSymlink(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()
	write(t, root, map[string]string{"a.txt": "a"})
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, link); err != nil {
		t.Skip("symbolic links:", err)
	}

	// A volume of a symbolic link is of the folder it's to, which is walked
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: link}, nil)
	require.NoError(t, err)
	assert.Equal(t, resolved(t, root), volume.Path)
	assert.Equal(t, []string{"a.txt"}, objects(t, manager, volume.ID))
	again, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, nil)
	require.NoError(t, err)
	assert.Equal(t, volume.ID, again.ID, "the same volume")
	assert.Equal(t, []string{"a.txt"}, objects(t, manager, volume.ID))
}

// resolved returns the path, of the folder which its symbolic links are to,
// such as /var, which is /private/var, on macOS
func resolved(t *testing.T, path string) string {
	t.Helper()
	path, err := filepath.EvalSymlinks(path)
	require.NoError(t, err)
	return path
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
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root, Recursive: true}, func(p schema.Progress, err error) {
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
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root, Recursive: true}, nil)
	require.NoError(t, err)
	paths := objects(t, manager, volume.ID)
	assert.Len(t, paths, 100, "a page of them")
	list, err := manager.ListObjects(ctx, schema.ObjectListRequest{Volumes: []int64{volume.ID}})
	require.NoError(t, err)
	assert.Equal(t, uint64(250), list.Count)
	_, err = manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root, Recursive: true}, nil)
	require.NoError(t, err)
	progress, err := manager.Progress(volume.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(250), progress.Marked)
	assert.True(t, sort.StringsAreSorted(paths))

	// A batch which is rolled back, as its context is done, isn't counted
	write(t, root, map[string]string{"files/new.txt": "x"})
	cancelled, cancel := context.WithCancel(ctx)
	_, err = manager.IndexVolume(cancelled, schema.IndexVolumeRequest{Path: root, Recursive: true}, func(p schema.Progress, err error) {
		if p.Indexed+p.Marked == 0 {
			return
		}
		cancel()
	})
	assert.ErrorIs(t, err, context.Canceled)
	progress, err = manager.Progress(volume.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(100), progress.Indexed+progress.Marked, "of the first batch, which was committed")
}

func Test_IndexVolumeSuperseded(t *testing.T) {
	ctx := t.Context()
	root := t.TempDir()
	write(t, root, map[string]string{"a.txt": "a", "b.txt": "b"})
	path := test.Path(t)
	a, err := filesearch.New(ctx, path)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, a.Close()) })
	b, err := filesearch.New(ctx, path)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, b.Close()) })

	// A volume, which is indexed
	volume, err := a.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, nil)
	require.NoError(t, err)

	// A run, of one manager, as of a process, which pauses once its files are
	// marked, and then another's, which pauses as it starts
	pause := func(m *filesearch.Manager) (chan struct{}, chan struct{}, chan error) {
		paused, resume, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
		var once sync.Once
		go func() {
			_, err := m.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, func(p schema.Progress, _ error) {
				if m == a && p.Marked == 0 {
					return
				}
				once.Do(func() {
					close(paused)
					<-resume
				})
			})
			done <- err
		}()
		<-paused
		return paused, resume, done
	}
	_, resumeA, doneA := pause(a)
	_, resumeB, doneB := pause(b)

	// The first, which is overtaken, doesn't finish the volume, or sweep its
	// files, which the second hasn't marked yet, and says so
	close(resumeA)
	assert.ErrorIs(t, <-doneA, sqlite.ErrConflict)
	assert.Equal(t, []string{"a.txt", "b.txt"}, objects(t, b, volume.ID), "nothing is swept")

	// The second finishes it
	close(resumeB)
	require.NoError(t, <-doneB)
	volumes, err := b.ListVolumes(ctx, schema.VolumeListRequest{})
	require.NoError(t, err)
	require.Len(t, volumes.Body, 1)
	assert.Equal(t, schema.StatusIndexed, volumes.Body[0].Status)
	assert.Equal(t, volume.Run+2, volumes.Body[0].Run)
	assert.Equal(t, []string{"a.txt", "b.txt"}, objects(t, b, volume.ID))
}
