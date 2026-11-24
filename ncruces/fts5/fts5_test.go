package fts5_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
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

// publisher keeps the tables of the changes which are published
type publisher struct{ tables []string }

func (p *publisher) Publish(change schema.Change) int {
	p.tables = append(p.tables, change.Schema+"."+change.Table)
	return 0
}

func Test_FTS5Changes(t *testing.T) {
	ctx := t.Context()

	// The changes of the shadow tables of a virtual table aren't published,
	// though those of the table which it indexes are
	p := &publisher{}
	db, err := ncruces.Open(ctx, t.TempDir()+"/test.db", sqlite.WithBroadcaster(p))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	require.NoError(t, db.Exec(ctx, `CREATE TABLE doc (id INTEGER PRIMARY KEY, body TEXT); CREATE VIRTUAL TABLE doc_text USING fts5(body)`))
	p.tables = nil
	require.NoError(t, db.Exec(ctx, `INSERT INTO doc (id, body) VALUES (1, 'rain'); INSERT INTO doc_text (rowid, body) VALUES (1, 'rain')`))
	assert.Equal(t, []string{"main.doc"}, p.tables)
}

func Test_FTS5ChangesSchemaName(t *testing.T) {
	ctx := t.Context()

	// With a schema name, the shadow tables' changes still aren't published,
	// as they're found by their database, and the table's are of the name
	p := &publisher{}
	db, err := ncruces.Open(ctx, t.TempDir()+"/test.db", sqlite.WithBroadcaster(p), sqlite.WithSchemaName("search"))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	require.NoError(t, db.Exec(ctx, `CREATE TABLE doc (id INTEGER PRIMARY KEY, body TEXT); CREATE VIRTUAL TABLE doc_text USING fts5(body)`))
	p.tables = nil
	require.NoError(t, db.Exec(ctx, `INSERT INTO doc (id, body) VALUES (1, 'rain'); INSERT INTO doc_text (rowid, body) VALUES (1, 'rain')`))
	assert.Equal(t, []string{"search.doc"}, p.tables)
}

func Test_FTS5ChangesRecreated(t *testing.T) {
	ctx := t.Context()

	// A table which is created, of the name of a shadow table which was
	// dropped, is published
	p := &publisher{}
	db, err := ncruces.Open(ctx, t.TempDir()+"/test.db", sqlite.WithBroadcaster(p))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	require.NoError(t, db.Exec(ctx, `CREATE VIRTUAL TABLE doc_text USING fts5(body)`))
	require.NoError(t, db.Exec(ctx, `INSERT INTO doc_text (rowid, body) VALUES (1, 'rain')`))
	require.NoError(t, db.Exec(ctx, `DROP TABLE doc_text; CREATE TABLE doc_text_data (id INTEGER PRIMARY KEY)`))
	p.tables = nil
	require.NoError(t, db.Exec(ctx, `INSERT INTO doc_text_data (id) VALUES (1)`))
	assert.Equal(t, []string{"main.doc_text_data"}, p.tables)
}
