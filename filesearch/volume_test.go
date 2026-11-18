package filesearch_test

import (
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
	test "github.com/mutablelogic/go-sqlite/filesearch/test"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_ListVolumes(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)

	// None, until they're indexed
	list, err := manager.ListVolumes(ctx, schema.VolumeListRequest{})
	require.NoError(t, err)
	assert.Zero(t, list.Count)
	assert.Empty(t, list.Body)

	// Volumes, by their paths, with the status of their indexing
	b, a := t.TempDir(), t.TempDir()
	_, err = manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: b}, nil)
	require.NoError(t, err)
	_, err = manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: a}, nil)
	require.NoError(t, err)
	list, err = manager.ListVolumes(ctx, schema.VolumeListRequest{})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), list.Count)
	if assert.Len(t, list.Body, 2) {
		assert.Less(t, list.Body[0].Path, list.Body[1].Path)
		for _, volume := range list.Body {
			assert.Equal(t, schema.StatusIndexed, volume.Status)
			assert.Equal(t, int64(1), volume.Run)
		}
	}
	assert.Equal(t, uint64(2), *list.Limit, "clamped to what's returned")

	// A page of them, and the caller's request isn't changed
	limit := uint64(5)
	list, err = manager.ListVolumes(ctx, schema.VolumeListRequest{OffsetLimit: sqlite.OffsetLimit{Offset: 1, Limit: &limit}})
	require.NoError(t, err)
	assert.Len(t, list.Body, 1)
	assert.Equal(t, uint64(1), *list.Limit)
	assert.Equal(t, uint64(5), limit)

}

func Test_Volumes(t *testing.T) {
	ctx := t.Context()
	manager := test.New(t)
	root := t.TempDir()
	write(t, root, map[string]string{"a.txt": "a", "notes/b.md": "# B\n\nbody"})
	volume, err := manager.IndexVolume(ctx, schema.IndexVolumeRequest{Path: root, Recursive: true}, nil)
	require.NoError(t, err)

	// A volume, by its id, and its path
	got, err := manager.GetVolume(ctx, volume.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), got.Objects)
	got, err = manager.VolumeByPath(ctx, root)
	require.NoError(t, err)
	assert.Equal(t, volume.ID, got.ID)
	_, err = manager.GetVolume(ctx, volume.ID+1)
	assert.ErrorIs(t, err, sqlite.ErrNotFound)
	_, err = manager.VolumeByPath(ctx, t.TempDir())
	assert.ErrorIs(t, err, sqlite.ErrNotFound)

	// It's deleted, with its objects, and then isn't
	deleted, err := manager.DeleteVolume(ctx, volume.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), deleted.Objects)
	list, err := manager.ListObjects(ctx, schema.ObjectListRequest{})
	require.NoError(t, err)
	assert.Zero(t, list.Count)
	_, err = manager.Progress(volume.ID)
	assert.ErrorIs(t, err, sqlite.ErrNotFound, "its progress isn't kept")
	_, err = manager.DeleteVolume(ctx, volume.ID)
	assert.ErrorIs(t, err, sqlite.ErrNotFound)
}
