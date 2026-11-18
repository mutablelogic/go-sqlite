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
