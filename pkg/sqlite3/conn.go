package sqlite3

import (
	// Modules
	"fmt"
	"runtime"
	"sync"

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
	sync.Mutex
	*sqlite3.ConnEx
	c chan struct{}
}

type ExecFunc sqlite3.ExecFunc

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func OpenPath(path string, flags sqlite3.OpenFlags) (*Conn, error) {
	poolconn := new(Conn)

	// Open database with flags
	if conn, err := sqlite3.OpenPathEx(path, flags, ""); err != nil {
		return nil, err
	} else {
		poolconn.ConnEx = conn
		poolconn.c = make(chan struct{})
	}

	// Finalizer to panic when connection not closed before garbage collection
	_, file, line, _ := runtime.Caller(1)
	runtime.SetFinalizer(poolconn, func(conn *Conn) {
		if conn.c != nil {
			panic(fmt.Sprintf("%s:%d: missing associated call to Close()", file, line))
		}
	})

	// Return success
	return poolconn, nil
}

func (conn *Conn) Close() error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()

	// Release resources
	close(conn.c)
	conn.c = nil

	// Close underlying connection
	return conn.ConnEx.Close()
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Execute SQL statement without preparing, and invoke a callback for each row of results
// which may return true to abort
func (conn *Conn) Exec(st SQStatement, fn ExecFunc) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()

	if st == nil {
		return ErrBadParameter.With("Exec")
	}
	return conn.ConnEx.Exec(st.Query(), sqlite3.ExecFunc(fn))
}

// Attach database as schema. If path is empty then a new in-memory database
// is attached.
func (conn *Conn) Attach(schema, path string) error {
	if schema == "" {
		return ErrBadParameter.Withf("%q", schema)
	}
	if path == "" {
		return conn.Attach(schema, defaultMemory)
	}
	return conn.Exec(Q("ATTACH DATABASE ", DoubleQuote(path), " AS ", QuoteIdentifier(schema)), nil)
}

// Detach named database as schema
func (conn *Conn) Detach(schema string) error {
	return conn.Exec(Q("DETACH DATABASE ", QuoteIdentifier(schema)), nil)
}
