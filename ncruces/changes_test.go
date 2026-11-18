package ncruces_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	sqlite "github.com/mutablelogic/go-sqlite"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// publisher keeps the changes which are published, and calls a function of
// each, when it's set
type publisher struct {
	sync.Mutex
	changes []schema.Change
	fn      func(schema.Change)
}

func (p *publisher) Publish(change schema.Change) int {
	p.Lock()
	p.changes = append(p.changes, change)
	fn := p.fn
	p.Unlock()
	if fn != nil {
		fn(change)
	}
	return 0
}

// take returns the changes which are published, as "action table row", and
// forgets them
func (p *publisher) take() []string {
	p.Lock()
	defer p.Unlock()
	result := make([]string, 0, len(p.changes))
	for _, change := range p.changes {
		result = append(result, strings.TrimSpace(fmt.Sprintf("%s %s.%s %s", change.Action, change.Schema, change.Table, change.Row)))
	}
	p.changes = nil
	return result
}

// openChanges opens a database, of a file, with a table, a and b, whose
// changes are published, of the tables, or all of them
func openChanges(t *testing.T, tables ...string) (sqlite.PoolConn, *publisher) {
	t.Helper()
	p := &publisher{}
	db, err := ncruces.Open(t.Context(), filepath.Join(t.TempDir(), "changes.db"), sqlite.WithBroadcaster(p, tables...))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	require.NoError(t, db.Exec(t.Context(), "CREATE TABLE a (id INTEGER PRIMARY KEY, name TEXT); CREATE TABLE b (id INTEGER PRIMARY KEY, name TEXT)"))
	assert.Empty(t, p.take(), "creating tables isn't a change of rows")
	return db, p
}

func Test_Changes(t *testing.T) {
	ctx := t.Context()
	db, p := openChanges(t)

	// A row which is written, of its action, database, table and rowid
	require.NoError(t, db.Exec(ctx, "INSERT INTO a (id, name) VALUES (1, 'one')"))
	assert.Equal(t, []string{"create main.a 1"}, p.take())
	require.NoError(t, db.Exec(ctx, "UPDATE a SET name = 'uno' WHERE id = 1"))
	assert.Equal(t, []string{"update main.a 1"}, p.take())
	require.NoError(t, db.Exec(ctx, "DELETE FROM a WHERE id = 1"))
	assert.Equal(t, []string{"delete main.a 1"}, p.take())

	// Rows which are written in a transaction, once it's committed, in the
	// order of their first writes, and each once
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Exec(ctx, "INSERT INTO a (id, name) VALUES (2, 'two'), (3, 'three')"))
		require.NoError(t, conn.Exec(ctx, "INSERT INTO b (id, name) VALUES (1, 'one')"))
		require.NoError(t, conn.Exec(ctx, "UPDATE a SET name = 'deux' WHERE id = 2"))
		assert.Empty(t, p.take(), "not before it's committed")
		return nil
	}))
	assert.Equal(t, []string{"create main.a 2", "create main.a 3", "create main.b 1"}, p.take())

	// Created, and deleted, which is nothing, and updated, and deleted,
	// which is deleted
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Exec(ctx, "INSERT INTO a (id, name) VALUES (4, 'four')"))
		require.NoError(t, conn.Exec(ctx, "DELETE FROM a WHERE id = 4"))
		require.NoError(t, conn.Exec(ctx, "UPDATE a SET name = 'trois' WHERE id = 3"))
		require.NoError(t, conn.Exec(ctx, "DELETE FROM a WHERE id = 3"))
		return nil
	}))
	assert.Equal(t, []string{"delete main.a 3"}, p.take())

	// Deleted, and created again, which is updated, as is a row which is
	// replaced
	require.NoError(t, db.Exec(ctx, "DELETE FROM a WHERE id = 2; INSERT INTO a (id, name) VALUES (2, 'two')"))
	assert.Equal(t, []string{"update main.a 2"}, p.take())
	require.NoError(t, db.Exec(ctx, "INSERT OR REPLACE INTO a (id, name) VALUES (2, 'deux')"))
	assert.Equal(t, []string{"update main.a 2"}, p.take())

	// A row whose rowid is updated is deleted, and created
	require.NoError(t, db.Exec(ctx, "UPDATE a SET id = 5 WHERE id = 2"))
	assert.Equal(t, []string{"delete main.a 2", "create main.a 5"}, p.take())

	// Each row which is deleted, without a WHERE
	require.NoError(t, db.Exec(ctx, "INSERT INTO b (id, name) VALUES (2, 'two')"))
	p.take()
	require.NoError(t, db.Exec(ctx, "DELETE FROM b"))
	assert.Equal(t, []string{"delete main.b 1", "delete main.b 2"}, p.take())
}

func Test_ChangesRolledBack(t *testing.T) {
	ctx := t.Context()
	db, p := openChanges(t)
	failed := errors.New("failed")

	// A transaction which is rolled back has no changes
	assert.ErrorIs(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Exec(ctx, "INSERT INTO a (id, name) VALUES (1, 'one')"))
		return failed
	}), failed)
	assert.Empty(t, p.take())

	// Nor does a savepoint which is rolled back to, of those of the
	// transaction
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		require.NoError(t, conn.Exec(ctx, "INSERT INTO a (id, name) VALUES (1, 'one')"))
		assert.ErrorIs(t, conn.Tx(ctx, func(conn sqlite.Conn) error {
			require.NoError(t, conn.Exec(ctx, "INSERT INTO a (id, name) VALUES (2, 'two')"))
			return failed
		}), failed)
		require.NoError(t, conn.Tx(ctx, func(conn sqlite.Conn) error {
			return conn.Exec(ctx, "INSERT INTO b (id, name) VALUES (1, 'one')")
		}))
		return nil
	}))
	assert.Equal(t, []string{"create main.a 1", "create main.b 1"}, p.take())

	// Nor a query of several statements, one of which fails, in a
	// transaction, or not
	assert.Error(t, db.Exec(ctx, "INSERT INTO a (id, name) VALUES (3, 'three'); INSERT INTO a (id, name) VALUES (1, 'duplicate')"))
	assert.Empty(t, p.take())
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		assert.Error(t, conn.Exec(ctx, "INSERT INTO a (id, name) VALUES (3, 'three'); INSERT INTO a (id, name) VALUES (1, 'duplicate')"))
		return conn.Exec(ctx, "INSERT INTO a (id, name) VALUES (4, 'four')")
	}))
	assert.Equal(t, []string{"create main.a 4"}, p.take())

	// Nor one which controls its own transaction, which is rolled back
	require.NoError(t, db.Exec(ctx, "BEGIN; INSERT INTO a (id, name) VALUES (5, 'five'); ROLLBACK"))
	assert.Empty(t, p.take())
}

func Test_ChangesTables(t *testing.T) {
	ctx := t.Context()
	db, p := openChanges(t, "B")

	// Only the tables' changes, of any case
	require.NoError(t, db.Exec(ctx, "INSERT INTO a (id, name) VALUES (1, 'one'); INSERT INTO b (id, name) VALUES (1, 'one')"))
	assert.Equal(t, []string{"create main.b 1"}, p.take())

	// Which are identifiers
	_, err := ncruces.Open(ctx, sqlite.Memory, sqlite.WithBroadcaster(p, "not a table"))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
	_, err = ncruces.Open(ctx, sqlite.Memory, sqlite.WithBroadcaster(nil))
	assert.ErrorIs(t, err, sqlite.ErrBadParameter)
}

func Test_ChangesMany(t *testing.T) {
	ctx := t.Context()
	db, p := openChanges(t)

	// More rows of a table than MaxChanges, of a transaction, are one change
	// of the table, and the others' are of their rows
	require.NoError(t, db.Exec(ctx, fmt.Sprintf(`
		WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n WHERE i < %d)
		INSERT INTO a (id, name) SELECT i, 'row' FROM n;
		INSERT INTO b (id, name) VALUES (1, 'one')`, sqlite.MaxChanges+1)))
	assert.Equal(t, []string{"create main.a", "create main.b 1"}, p.take())
}

func Test_ChangesCommitted(t *testing.T) {
	ctx := t.Context()
	db, p := openChanges(t)

	// A change is published once it's committed, so a watcher which reads it,
	// on another goroutine, reads it
	var wg sync.WaitGroup
	var names []string
	p.fn = func(change schema.Change) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var v struct{ name string }
			err := db.Get(context.WithoutCancel(ctx), readerFunc(func(row sqlite.Row) error { return row.Scan(&v.name) }), selectFunc("SELECT name FROM a WHERE id = "+change.Row))
			assert.NoError(t, err)
			names = append(names, v.name)
		}()
	}
	require.NoError(t, db.Tx(ctx, func(conn sqlite.Conn) error {
		return conn.Exec(ctx, "INSERT INTO a (id, name) VALUES (1, 'one')")
	}))
	wg.Wait()
	assert.Equal(t, []string{"one"}, names)
}

// readerFunc scans a row
type readerFunc func(sqlite.Row) error

func (fn readerFunc) Scan(row sqlite.Row) error { return fn(row) }

// selectFunc is a query, which selects
type selectFunc string

func (q selectFunc) Select(*sqlite.Bind, sqlite.Op) (string, error) { return string(q), nil }

func Test_ChangesCascade(t *testing.T) {
	ctx := t.Context()
	db, p := openChanges(t)

	// The rows which a foreign key deletes, as their parent is
	require.NoError(t, db.Exec(ctx, "CREATE TABLE c (id INTEGER PRIMARY KEY, a INTEGER REFERENCES a(id) ON DELETE CASCADE)"))
	require.NoError(t, db.Exec(ctx, "INSERT INTO a (id, name) VALUES (1, 'one'); INSERT INTO c (id, a) VALUES (1, 1), (2, 1)"))
	p.take()
	require.NoError(t, db.Exec(ctx, "DELETE FROM a WHERE id = 1"))
	assert.ElementsMatch(t, []string{"delete main.a 1", "delete main.c 1", "delete main.c 2"}, p.take())
}
