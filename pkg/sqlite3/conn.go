package sqlite3

import (
	"context"
	"fmt"
	"os"
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
	c   chan struct{}
	ctx context.Context
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

	// If no create flag then check to make sure database exists
	if path != defaultMemory && flags&sqlite3.SQLITE_OPEN_CREATE == 0 {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, ErrNotFound.Withf("%q", path)
		} else if err != nil {
			return nil, err
		}
	}

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
func (conn *Conn) Do(ctx context.Context, flag SQTxnFlag, fn func(SQTransaction) error) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()

	// Return any context errors
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}

	// Get existing foreign key constraints, set new ones
	fk, err := conn.ForeignKeyConstraints()
	if err != nil {
		return err
	}
	if flag&SQLITE_TXN_NO_FOREIGNKEY_CONSTRAINTS != 0 && fk {
		if err := conn.SetForeignKeyConstraints(false); err != nil {
			return err
		}
	}

	// Flags
	v := sqlite3.SQLITE_TXN_DEFAULT
	if flag.Is(SQLITE_TXN_EXCLUSIVE) {
		v = sqlite3.SQLITE_TXN_EXCLUSIVE
	} else if flag.Is(SQLITE_TXN_IMMEDIATE) {
		v = sqlite3.SQLITE_TXN_IMMEDIATE
	}

	// Begin transaction
	if err := conn.ConnEx.Begin(v); err != nil {
		return err
	}

	// Perform transaction
	var result error
	if fn != nil {
		conn.ctx = ctx
		conn.SetProgressHandler(100, func() bool {
			return ctx != nil && ctx.Err() != nil
		})
		if err := fn(&Txn{conn}); err != nil {
			result = multierror.Append(result, err)
		}
		conn.SetProgressHandler(0, nil)
		conn.ctx = nil
	}

	// Commit or rollback transaction
	if result == nil {
		result = multierror.Append(result, conn.ConnEx.Commit())
	} else {
		result = multierror.Append(result, conn.ConnEx.Rollback())
	}

	// Return foreign key constraints
	result = multierror.Append(result, conn.SetForeignKeyConstraints(fk))

	// Return any errors
	return result
}

// Execute SQL statement and invoke a callback for each row of results which may return true to abort
func (txn *Txn) Query(st SQStatement, v ...interface{}) (SQResult, error) {
	if st == nil {
		return nil, ErrBadParameter.With("Query")
	}

	// TODO: Get prepared version of query from the cache
	stx, err := txn.Conn.Prepare(st.Query())
	if err != nil {
		return nil, err
	}

	// Create a results object and return it
	results, err := stx.Exec(0, v...)
	if err != nil {
		stx.Close()
		return nil, err
	}

	return NewResults(results), nil
}
