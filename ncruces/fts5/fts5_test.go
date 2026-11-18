package fts5_test

import (
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
	_ "github.com/mutablelogic/go-sqlite/ncruces/fts5"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// match is the rowid of a match
type match struct {
	query string
	rowid int64
}

func (m *match) Select(*sqlite.Bind, sqlite.Op) (string, error) { return m.query, nil }
func (m *match) Scan(row sqlite.Row) error                      { return row.Scan(&m.rowid) }

func Test_FTS5(t *testing.T) {
	ctx := t.Context()

	// Full-text search, on the writer, and the readers, of a file
	db, err := ncruces.Open(ctx, t.TempDir()+"/test.db")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	require.NoError(t, db.Exec(ctx, `CREATE VIRTUAL TABLE docs USING fts5(title, body)`))
	require.NoError(t, db.Exec(ctx, `INSERT INTO docs (rowid, title, body) VALUES (1, 'Blade Runner', 'replicants in the rain'), (2, 'Alien', 'in space no one can hear you scream')`))
	m := &match{query: `SELECT rowid FROM docs WHERE docs MATCH 'rain' ORDER BY rank`}
	require.NoError(t, db.Get(ctx, m, m))
	assert.Equal(t, int64(1), m.rowid)
	m = &match{query: `SELECT rowid FROM docs WHERE docs MATCH 'title:alien'`}
	require.NoError(t, db.Get(ctx, m, m))
	assert.Equal(t, int64(2), m.rowid)
}
