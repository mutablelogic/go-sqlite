package ncruces_test

import (
	"context"
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
	test "github.com/mutablelogic/go-sqlite/test"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_Ncruces(t *testing.T) {
	test.Run(t, ncruces.Open)
}

// value is a query's one integer
type value struct {
	query string
	value int64
}

func (v *value) Select(*sqlite.Bind, sqlite.Op) (string, error) { return v.query, nil }
func (v *value) Scan(row sqlite.Row) error                      { return row.Scan(&v.value) }

func Test_NoExtensions(t *testing.T) {
	// Without the subfolders of extensions imported, there are none
	db, err := ncruces.Open(t.Context(), sqlite.Memory)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	assert.ErrorIs(t, db.Exec(t.Context(), "CREATE VIRTUAL TABLE f USING fts5(body)"), sqlite.ErrDatabase)
}

// row is inserted, and ignored
type row struct{ name string }

func (r row) Insert(bind *sqlite.Bind) (string, error) {
	bind.Set("name", r.name)
	return `INSERT INTO t (name) VALUES (@name)`, nil
}
func (row) Update(*sqlite.Bind) error { return nil }

func Benchmark_Insert(b *testing.B) {
	ctx := b.Context()
	db, err := ncruces.Open(ctx, b.TempDir()+"/test.db")
	require.NoError(b, err)
	b.Cleanup(func() { _ = db.Close() })
	require.NoError(b, db.Exec(ctx, `CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT)`))

	// 1000 rows, in a transaction, and in a bulk operation, which reuses its
	// statement
	for _, bench := range []struct {
		name string
		fn   func(context.Context, func(sqlite.Conn) error) error
	}{{"Tx", db.Tx}, {"Bulk", db.Bulk}} {
		b.Run(bench.name, func(b *testing.B) {
			for b.Loop() {
				require.NoError(b, bench.fn(ctx, func(conn sqlite.Conn) error {
					for range 1000 {
						if err := conn.Insert(ctx, nil, row{"name"}); err != nil {
							return err
						}
					}
					return nil
				}))
			}
		})
	}
}
