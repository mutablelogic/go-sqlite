package ncruces_test

import (
	"context"
	"errors"
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
	test "github.com/mutablelogic/go-sqlite/test"
	sqlite3 "github.com/ncruces/go-sqlite3"
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

func Test_WithInit(t *testing.T) {
	ctx := t.Context()

	// A function, of ncruces's, which each connection has, the readers' too
	double := func(conn *sqlite3.Conn) error {
		return conn.CreateFunction("double", 1, sqlite3.DETERMINISTIC, func(ctx sqlite3.Context, arg ...sqlite3.Value) {
			ctx.ResultInt64(arg[0].Int64() * 2)
		})
	}
	db, err := ncruces.Open(ctx, t.TempDir()+"/test.db", ncruces.WithInit(double, nil))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	v := &value{query: "SELECT double(21)"}
	require.NoError(t, db.Get(ctx, v, v))
	assert.Equal(t, int64(42), v.value)

	// A function which fails, fails to open the database
	failed := errors.New("failed")
	_, err = ncruces.Open(ctx, sqlite.Memory, ncruces.WithInit(func(*sqlite3.Conn) error { return failed }))
	assert.ErrorIs(t, err, failed)
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
