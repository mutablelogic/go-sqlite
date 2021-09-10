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
