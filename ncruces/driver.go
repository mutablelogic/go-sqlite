// Package ncruces is a driver for SQLite, of ncruces/go-sqlite3, which is
// SQLite compiled to WebAssembly, so it's pure Go, without cgo.
package ncruces

import (
	"database/sql"
	"errors"
	"net/url"
	"strings"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	sqlite3 "github.com/ncruces/go-sqlite3"
	driver "github.com/ncruces/go-sqlite3/driver"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Driver opens SQLite databases with ncruces/go-sqlite3
type Driver struct {
	init []func(*sqlite3.Conn) error
}

// Opt is an option of the driver
type Opt func(*Driver)

// Ensure interfaces are satisfied
var _ sqlite.Driver = (*Driver)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New returns a driver, with its options
func New(opts ...Opt) *Driver {
	d := new(Driver)
	for _, opt := range opts {
		opt(d)
	}
	return d
}

////////////////////////////////////////////////////////////////////////////////
// OPTIONS

// WithInit calls the functions, in order, on each connection as it's opened,
// such as to create functions, collations or virtual tables, or to register
// an extension, of github.com/ncruces/go-sqlite3/ext, whose Register
// functions are of the same signature.
func WithInit(fn ...func(*sqlite3.Conn) error) Opt {
	return func(d *Driver) {
		for _, fn := range fn {
			if fn != nil {
				d.init = append(d.init, fn)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Open opens the database of the path, or ":memory:", with the pragmas set on
// each connection. A writer's transactions begin immediately, and a reader's
// when they read.
func (d *Driver) Open(path string, readonly bool, pragmas []sqlite.Pragma) (*sql.DB, error) {
	query := url.Values{}
	for _, pragma := range pragmas {
		query.Add("_pragma", pragma.Name+"("+pragma.Value+")")
	}
	if readonly {
		query.Set("_txlock", "deferred")
	} else {
		query.Set("_txlock", "immediate")
	}
	return driver.Open(dsn(path)+"?"+query.Encode(), d.connect)
}

// Error returns SQLite's error as a DatabaseError, of the kinds of error it
// is, or the error as it is, when it isn't SQLite's
func (d *Driver) Error(err error) error {
	var sqliteErr *sqlite3.Error
	var code sqlite3.ErrorCode
	switch {
	case errors.As(err, &sqliteErr):
		return sqlite.NewDatabaseError(err, kinds(sqliteErr.Code(), sqliteErr.ExtendedCode())...)
	case errors.As(err, &code):
		return sqlite.NewDatabaseError(err, kinds(code, 0)...)
	}
	return err
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// connect calls the functions on a connection, as it's opened
func (d *Driver) connect(conn *sqlite3.Conn) error {
	for _, fn := range d.init {
		if err := fn(conn); err != nil {
			return err
		}
	}
	return nil
}

// dsn returns the URI of the path, whose characters, such as ? and #, are
// escaped, so they're of its name
func dsn(path string) string {
	if path == sqlite.Memory {
		return "file::memory:"
	}
	return "file:" + strings.ReplaceAll((&url.URL{Path: path}).EscapedPath(), "?", "%3F")
}

// kinds returns the kinds of error of SQLite's codes
func kinds(code sqlite3.ErrorCode, extended sqlite3.ExtendedErrorCode) []sqlite.Err {
	switch extended {
	case sqlite3.CONSTRAINT_UNIQUE, sqlite3.CONSTRAINT_PRIMARYKEY:
		return []sqlite.Err{sqlite.ErrConflict, sqlite.ErrUniqueViolation}
	case sqlite3.CONSTRAINT_FOREIGNKEY:
		return []sqlite.Err{sqlite.ErrBadParameter, sqlite.ErrForeignKeyViolation}
	case sqlite3.CONSTRAINT_NOTNULL:
		return []sqlite.Err{sqlite.ErrBadParameter, sqlite.ErrNotNullViolation}
	case sqlite3.CONSTRAINT_CHECK:
		return []sqlite.Err{sqlite.ErrBadParameter, sqlite.ErrCheckViolation}
	}
	switch code {
	case sqlite3.CONSTRAINT:
		return []sqlite.Err{sqlite.ErrBadParameter}
	case sqlite3.BUSY, sqlite3.LOCKED:
		return []sqlite.Err{sqlite.ErrNotAvailable}
	case sqlite3.READONLY:
		return []sqlite.Err{sqlite.ErrBadParameter}
	case sqlite3.CANTOPEN, sqlite3.NOTADB:
		return []sqlite.Err{sqlite.ErrNotAvailable}
	}
	return nil
}
