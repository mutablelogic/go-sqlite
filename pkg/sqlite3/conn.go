package sqlite3

import (
	"fmt"
	"runtime"
	"sync"

	// Modules
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"
	multierror "github.com/hashicorp/go-multierror"

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

type Txn struct {
	*Conn
}

type ExecFunc sqlite3.ExecFunc
type TxnFunc func(SQTransaction) error

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
	if st == nil {
		return ErrBadParameter.With("Exec")
	}
	return conn.ConnEx.Exec(st.Query(), sqlite3.ExecFunc(fn))
}

// Perform a transaction, rollback if error is returned
func (conn *Conn) Do(fn TxnFunc) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()

	// Begin transaction
	if err := conn.ConnEx.Begin(sqlite3.SQLITE_TXN_DEFAULT); err != nil {
		return err
	}

	var result error
	if fn != nil {
		if err := fn(&Txn{conn}); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if result == nil {
		result = multierror.Append(result, conn.ConnEx.Commit())
	} else {
		result = multierror.Append(result, conn.ConnEx.Rollback())
	}

	// Return any errors
	return nil
}

// Execute SQL statement and invoke a callback for each row of results which may return true to abort
func (txn *Txn) Query(st SQStatement, v ...interface{}) (SQResult, error) {
	if st == nil {
		return nil, ErrBadParameter.With("Query")
	}
	return nil, ErrNotImplemented
}
