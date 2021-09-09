package sqlite3

import (
	// Modules
	"github.com/djthorpe/go-sqlite/sys/sqlite3"

	// Namespace Imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
	. "github.com/djthorpe/go-sqlite/pkg/quote"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Conn struct {
	*sqlite3.ConnEx
}

type ExecFunc sqlite3.ExecFunc

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func OpenPath(path string, flags sqlite3.OpenFlags) (*Conn, error) {
	this := new(Conn)

	// Open database with flags
	if c, err := sqlite3.OpenPathEx(path, flags, ""); err != nil {
		return nil, err
	} else {
		this.ConnEx = c
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Execute SQL statement without preparing, and invoke a callback for each row of results
// which may return true to abort
func (this *Conn) Exec(st SQStatement, fn ExecFunc) error {
	if st == nil {
		return ErrBadParameter.With("Exec")
	}
	return this.ConnEx.Exec(st.Query(), sqlite3.ExecFunc(fn))
}

// Attach database as schema. If path is empty then a new in-memory database
// is attached.
func (this *Conn) Attach(schema, path string) error {
	if schema == "" {
		return ErrBadParameter.Withf("%q", schema)
	}
	if path == "" {
		return this.Attach(schema, defaultMemory)
	}
	return this.Exec(Q("ATTACH DATABASE ", DoubleQuote(path), " AS ", QuoteIdentifier(schema)), nil)
}

// Detach named database as schema
func (this *Conn) Detach(schema string) error {
	return this.Exec(Q("DETACH DATABASE ", QuoteIdentifier(schema)), nil)
}
