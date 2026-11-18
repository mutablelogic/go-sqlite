package rtree_test

import (
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
	_ "github.com/mutablelogic/go-sqlite/ncruces/rtree"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// found is the id of a box which is found
type found struct {
	query string
	id    int64
}

func (f *found) Select(*sqlite.Bind, sqlite.Op) (string, error) { return f.query, nil }
func (f *found) Scan(row sqlite.Row) error                      { return row.Scan(&f.id) }

func Test_RTree(t *testing.T) {
	ctx := t.Context()

	// Boxes, which are found by where they are
	db, err := ncruces.Open(ctx, t.TempDir()+"/test.db")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	require.NoError(t, db.Exec(ctx, `CREATE VIRTUAL TABLE places USING rtree(id, minx, maxx, miny, maxy)`))
	require.NoError(t, db.Exec(ctx, `INSERT INTO places VALUES (1, 0, 1, 0, 1), (2, 10, 11, 10, 11)`))
	f := &found{query: `SELECT id FROM places WHERE minx >= 9 AND maxx <= 12 AND miny >= 9 AND maxy <= 12`}
	require.NoError(t, db.Get(ctx, f, f))
	assert.Equal(t, int64(2), f.id)
}
