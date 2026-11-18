package vec1_test

import (
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
	_ "github.com/mutablelogic/go-sqlite/ncruces/vec1"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// count is a query's count
type count struct {
	query string
	n     int64
}

func (c *count) Select(*sqlite.Bind, sqlite.Op) (string, error) { return c.query, nil }
func (c *count) Scan(row sqlite.Row) error                      { return row.Scan(&c.n) }

func Test_Vec1(t *testing.T) {
	ctx := t.Context()

	// Vectors, of an index, which are stored, and counted
	db, err := ncruces.Open(ctx, t.TempDir()+"/test.db")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	require.NoError(t, db.Exec(ctx, `CREATE VIRTUAL TABLE v USING vec1`))
	require.NoError(t, db.Exec(ctx, `INSERT INTO v (cmd, vector) VALUES ('rebuild', '{index:"flat"}')`))
	require.NoError(t, db.With("vector", "[1,2,3,4,5,6,7,8]").Exec(ctx, `INSERT INTO v (rowid, vector) VALUES (1, vec1_from_json(@vector))`))
	c := &count{query: `SELECT COUNT(*) FROM v`}
	require.NoError(t, db.Get(ctx, c, c))
	assert.Equal(t, int64(1), c.n)
}
