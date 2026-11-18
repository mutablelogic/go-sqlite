package filesearch_test

import (
	"os"
	"path/filepath"
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	filesearch "github.com/mutablelogic/go-sqlite/filesearch"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
	test "github.com/mutablelogic/go-sqlite/filesearch/test"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// excludeFiles are files, and folders, which are excluded by default, or
// aren't
var excludeFiles = map[string]string{
	"a.txt":                        "a",
	".hidden":                      "h",
	".git/config":                  "g",
	"notes/.DS_Store":              "d",
	"notes/b.md":                   "b",
	"notes/~$report.docx":          "l",
	"notes/report.docx~":           "l",
	"notes/#draft.md#":             "l",
	"downloads/film.mp4.part":      "p",
	"downloads/setup.CRDOWNLOAD":   "p",
	"Thumbs.db":                    "t",
	"code/node_modules/x/index.js": "n",
	"code/main.py":                 "m",
	"code/__pycache__/main.pyc":    "c",
	"code/old.pyc":                 "c",
	"backup/c.bak":                 "k",
}

func Test_Exclude(t *testing.T) {
	ctx := t.Context()
	root := t.TempDir()
	write(t, root, excludeFiles)

	// By default, hidden files and folders, lock files, backups, temporary
	// files, systems' files, and dependencies, aren't indexed, and others,
	// of WithExclude, too
	manager := test.New(t, filesearch.WithExclude(`\.bak$`))
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root, Recursive: true}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"a.txt", "code/main.py", "notes/b.md"}, objects(t, manager, volume.ID))
	progress, err := manager.Progress(volume.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), progress.Files, "what's excluded isn't counted")

	// Without the defaults, they're indexed, but what WithExclude excludes
	all := test.New(t, filesearch.WithoutDefaultExclude(), filesearch.WithExclude(`\.bak$`))
	volume, err = all.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root, Recursive: true}, nil)
	require.NoError(t, err)
	assert.Len(t, objects(t, all, volume.ID), len(excludeFiles)-1)

	// A volume of a folder, which would be excluded, is indexed
	hidden := test.New(t)
	volume, err = hidden.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root + "/.git"}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"config"}, objects(t, hidden, volume.ID))

	// A regular expression which isn't is a bad parameter
	_, err = filesearch.New(ctx, sqlite.Memory, filesearch.WithExclude(`(`))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
}

func Test_ExcludeSweeps(t *testing.T) {
	ctx := t.Context()
	root := t.TempDir()
	write(t, root, map[string]string{"a.txt": "a", "b.bak": "b"})
	path := test.Path(t)

	// What's indexed, and then excluded, is swept, when it's indexed again
	manager, err := filesearch.New(ctx, path)
	require.NoError(t, err)
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"a.txt", "b.bak"}, objects(t, manager, volume.ID))
	require.NoError(t, manager.Close())
	manager, err = filesearch.New(ctx, path, filesearch.WithExclude(`\.bak$`))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, manager.Close()) })
	_, err = manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"a.txt"}, objects(t, manager, volume.ID))
}

func Test_ExcludeUnreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permissions aren't enforced")
	}
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()
	write(t, root, map[string]string{"a.txt": "a", ".private/b.txt": "b"})
	private := filepath.Join(root, ".private")
	require.NoError(t, os.Chmod(private, 0o000))
	t.Cleanup(func() { _ = os.Chmod(private, 0o755) })

	// A folder which is excluded isn't an error, though it can't be read
	_, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root, Recursive: true}, func(p schema.Progress, err error) {
		assert.NoError(t, err)
	})
	require.NoError(t, err)
}
