package ncruces_test

import (
	"errors"
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
	sqlitetest "github.com/mutablelogic/go-sqlite/sqlitetest"
	sqlite3 "github.com/ncruces/go-sqlite3"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_Driver(t *testing.T) {
	sqlitetest.Run(t, ncruces.New())
}

// value is a query's one value
type value struct {
	query string
	value int64
}

func (v *value) Select(*sqlite.Bind, sqlite.Op) (string, error) { return v.query, nil }
func (v *value) Scan(row sqlite.Row) error                      { return row.Scan(&v.value) }

func Test_WithInit(t *testing.T) {
	ctx := t.Context()

	// A function, which each connection has, the readers' too
	double := func(conn *sqlite3.Conn) error {
		return conn.CreateFunction("double", 1, sqlite3.DETERMINISTIC, func(ctx sqlite3.Context, arg ...sqlite3.Value) {
			ctx.ResultInt64(arg[0].Int64() * 2)
		})
	}
	db, err := sqlite.Open(ctx, t.TempDir()+"/test.db", ncruces.New(ncruces.WithInit(double, nil)))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	v := &value{query: "SELECT double(21)"}
	require.NoError(t, db.Get(ctx, v, v))
	assert.Equal(t, int64(42), v.value)

	// A function which fails, fails to open the database
	failed := errors.New("failed")
	_, err = sqlite.Open(ctx, sqlite.Memory, ncruces.New(ncruces.WithInit(func(*sqlite3.Conn) error { return failed })))
	assert.ErrorIs(t, err, failed)
}
