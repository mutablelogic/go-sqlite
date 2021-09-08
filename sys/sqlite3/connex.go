package sqlite3

/*
#cgo pkg-config: sqlite3
#include <sqlite3.h>
#include <stdlib.h>

extern int go_busy_handler(void* userInfo, int n);
static inline int _sqlite3_busy_handler(sqlite3* db, uintptr_t userInfo) {
	return sqlite3_busy_handler(db, go_busy_handler, (void* )(userInfo));
}

extern int go_progress_handler(void* userInfo);
static inline void _sqlite3_progress_handler(sqlite3* db, int n, uintptr_t userInfo) {
	sqlite3_progress_handler(db, n, go_progress_handler, (void* )(userInfo));
}

extern int go_commit_hook(void* userInfo);
static inline void _sqlite3_commit_hook(sqlite3* db, uintptr_t userInfo) {
	sqlite3_commit_hook(db, go_commit_hook, (void* )(userInfo));
}

extern void go_rollback_hook(void* userInfo);
static inline void _sqlite3_rollback_hook(sqlite3* db, uintptr_t userInfo) {
	sqlite3_rollback_hook(db, go_rollback_hook, (void* )(userInfo));
}

extern void go_update_hook(void* userInfo, int op, char* db, char* tbl, sqlite3_int64 row);
static inline void _sqlite3_update_hook(sqlite3* db, uintptr_t userInfo) {
	sqlite3_update_hook(db, (void (*)(void* , int, char const*,char const*, sqlite3_int64))(go_update_hook), (void* )(userInfo));
}

extern int go_authorizer_hook(void* userInfo, int op, char* a1, char* a2, char* a3, char* a4);
static inline void _sqlite3_set_authorizer(sqlite3* db, uintptr_t userInfo) {
	sqlite3_set_authorizer(db, (int (*)(void*, int, const char*, const char*, const char*, const char*))(go_authorizer_hook), (void*)(userInfo));
}

extern int go_exec_handler(void* userInfo, int nargs, char** row, char** cols);
static inline int _sqlite3_exec(sqlite3* db, char* q, uintptr_t userInfo, char** errmsg) {
	return sqlite3_exec(db, (const char* )(q), go_exec_handler, (void* )(userInfo), errmsg);
}

*/
import "C"

import (
	"errors"
	"reflect"
	"sync"
	"time"
	"unsafe"

	// Modules
	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ConnEx struct {
	*Conn

	// Callback functions
	BusyHandlerFunc
	ProgressHandlerFunc
	CommitHookFunc
	RollbackHookFunc
	UpdateHookFunc
	AuthorizerHookFunc
	ExecFunc

	// Locks
	xmu sync.Mutex // Mutex for calling Exec - only one per connection
}

// BusyHandlerFunc is invoked with the number of times that the busy handler has been invoked previously
// for the same locking event. If the busy callback returns false, then no additional attempts are
// made to access the database and error SQLITE_BUSY is returned to the application. If the callback
// returns true then another attempt is made to access the database and the cycle repeats.
type BusyHandlerFunc func(int) bool

// ProgressHandlerFunc is invoked periodically during long running calls. If the progress callback returns
// true, the operation is interrupted
type ProgressHandlerFunc func() bool

// CommitHookFunc is invoked on commit. When it returns false, the COMMIT operation is allowed to
// continue normally or else the COMMIT is converted into a ROLLBACK
type CommitHookFunc func() bool

// RollbackHookFunc is invoked whenever a transaction is rolled back
type RollbackHookFunc func()

// UpdateHookFunc is invoked whenever a row is updated, inserted or deleted
// SQOperation will be one of SQLITE_INSERT, SQLITE_DELETE, or SQLITE_UPDATE.
// The other arguments are database name, table name and the rowid of the row.
// In the case of an update, this is the rowid after the update takes place.
type UpdateHookFunc func(SQAction, string, string, int64)

// AuthorizerHookFunc is invoked as SQL statements are being compiled by sqlite3_prepare
// the arguments are dependent on the action required, and the return value should be
// SQLITE_ALLOW, SQLITE_DENY or SQLITE_IGNORE
type AuthorizerHookFunc func(SQAction, [4]string) SQAuth

// ExecFunc is invoked during an Exec call with row text values and column names.
// If an sqlite3_exec() callback returns true, the sqlite3_exec() routine returns
// SQLITE_ABORT without invoking the callback again and without running any subsequent
// SQL statements.
type ExecFunc func(row, cols []string) bool

// Transaction type
type SQTransaction string

// callback tracks ConnEx objects against userInfo data
type callback struct {
	sync.RWMutex
	fn map[uintptr]*ConnEx
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultBusyTimeout = 5 * time.Second
)

var (
	cb = callback{fn: make(map[uintptr]*ConnEx)}
)

const (
	SQLITE_TXN_DEFAULT   SQTransaction = "DEFERRED"
	SQLITE_TXN_IMMEDIATE SQTransaction = "IMMEDIATE"
	SQLITE_TXN_EXCLUSIVE SQTransaction = "EXCLUSIVE"
)

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Open URL (with busy and progress handlers)
func OpenUrlEx(url string, flags OpenFlags, vfs string) (*ConnEx, error) {
	return OpenPathEx(url, flags|SQLITE_OPEN_URI, vfs)
}

// Open Path (with busy and progress handlers)
func OpenPathEx(path string, flags OpenFlags, vfs string) (*ConnEx, error) {
	c := new(ConnEx)
	if conn, err := OpenPath(path, flags, vfs); err != nil {
		return nil, err
	} else {
		c.Conn = conn
	}

	// Add callback
	cb.add(c)

	// Set busy timeout
	if err := c.SetBusyTimeout(defaultBusyTimeout); err != nil {
		c.Conn.Close()
		return nil, err
	}

	// Return success
	return c, nil
}

// Close Connection
func (c *ConnEx) Close() error {
	var result error

	// Remove callbacks
	if err := c.SetBusyHandler(nil); err != nil {
		result = multierror.Append(result, err)
	}
	if err := c.SetProgressHandler(0, nil); err != nil {
		result = multierror.Append(result, err)
	}
	if err := c.SetCommitHook(nil); err != nil {
		result = multierror.Append(result, err)
	}
	if err := c.SetRollbackHook(nil); err != nil {
		result = multierror.Append(result, err)
	}
	if err := c.SetUpdateHook(nil); err != nil {
		result = multierror.Append(result, err)
	}
	if err := c.SetAuthorizerHook(nil); err != nil {
		result = multierror.Append(result, err)
	}

	// Remove callback from global var
	cb.delete(c)

	// Call close
	if err := c.Conn.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	// Return any errors
	return result
}

// Set Busy Timeout
func (c *ConnEx) SetBusyTimeout(t time.Duration) error {
	c.SetBusyHandler(nil)
	if err := SQError(C.sqlite3_busy_timeout((*C.sqlite3)(c.Conn), C.int(t/time.Millisecond))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Set Busy Handler, use nil to remove the handler
func (c *ConnEx) SetBusyHandler(fn BusyHandlerFunc) error {
	c.BusyHandlerFunc = fn

	// Add busy handler
	if err := SQError(C._sqlite3_busy_handler((*C.sqlite3)(c.Conn), C.uintptr_t(c.userInfo()))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Set Progress Handler, use nil to remove the handler. The parameter n is
// the approximate number of virtual machine instructions that are evaluated between
// successive invocations of the callback
func (c *ConnEx) SetProgressHandler(n uint, fn ProgressHandlerFunc) error {
	if fn == nil || n == 0 {
		c.ProgressHandlerFunc = nil
	} else {
		c.ProgressHandlerFunc = fn
	}

	// Add progress handler
	C._sqlite3_progress_handler((*C.sqlite3)(c.Conn), C.int(n), C.uintptr_t(c.userInfo()))

	// Return success
	return nil
}

// SetCommitHook sets the callback for the commit hook, use nil to remove the handler.
func (c *ConnEx) SetCommitHook(fn CommitHookFunc) error {
	c.CommitHookFunc = fn

	// Add commit hook
	C._sqlite3_commit_hook((*C.sqlite3)(c.Conn), C.uintptr_t(c.userInfo()))

	// Return success
	return nil
}

// SetRollbackHook sets the callback for the rollback hook, use nil to remove the handler.
func (c *ConnEx) SetRollbackHook(fn RollbackHookFunc) error {
	c.RollbackHookFunc = fn

	// Add rollback hook
	C._sqlite3_rollback_hook((*C.sqlite3)(c.Conn), C.uintptr_t(c.userInfo()))

	// Return success
	return nil
}

// SetUpdateHook sets the callback for the update hook, use nil to remove the handler.
func (c *ConnEx) SetUpdateHook(fn UpdateHookFunc) error {
	c.UpdateHookFunc = fn

	// Add rollback hook
	C._sqlite3_update_hook((*C.sqlite3)(c.Conn), C.uintptr_t(c.userInfo()))

	// Return success
	return nil
}

// SetAuthorizerHook sets the callback for the authorizer hook, use nil to remove the handler.
func (c *ConnEx) SetAuthorizerHook(fn AuthorizerHookFunc) error {
	c.AuthorizerHookFunc = fn

	// Add rollback hook
	C._sqlite3_set_authorizer((*C.sqlite3)(c.Conn), C.uintptr_t(c.userInfo()))

	// Return success
	return nil
}

// Exec runs statements without preparing or accepting bind arguments. If you
// call with fn as nil then the statement is executed without a callback, otherwise
// return true from the callback to abort the transaction.
func (c *ConnEx) Exec(q string, fn ExecFunc) error {
	c.xmu.Lock()
	defer c.xmu.Unlock()

	// Set exec callback
	c.ExecFunc = fn
	defer func() {
		c.ExecFunc = nil
	}()

	// Populate CStrings
	var cErrmsg, cQuery *C.char
	cQuery = C.CString(q)
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cErrmsg))

	// Call exec
	var result error
	if err := SQError(C._sqlite3_exec((*C.sqlite3)(c.Conn), cQuery, C.uintptr_t(c.userInfo()), &cErrmsg)); err != SQLITE_OK {
		result = multierror.Append(result, err)
	}

	// Check errmsg
	if cErrmsg != nil {
		result = multierror.Append(result, errors.New(C.GoString(cErrmsg)))
		C.sqlite3_free(unsafe.Pointer(cErrmsg))
	}

	// Return any errors
	return result
}

// ExecEx runs statements without preparing. If you call with fn as nil then the
// statement is executed without a callback, otherwise callback is invoked for
// each row of data returned, otherwise return true from the callback to abort the
// transaction.
func (c *ConnEx) ExecEx(q string, fn ExecFunc, v ...interface{}) error {
	c.xmu.Lock()
	defer c.xmu.Unlock()

	// Prepare statements
	st, err := c.Prepare(q)
	if err != nil {
		return err
	}
	defer st.Close()

	// Execute loop
	for n := uint(0); ; n++ {
		r, err := st.Exec(n, v...)
		if err == SQLITE_DONE || r == nil {
			break
		} else if err != nil {
			return err
		}

		// Cast row values to string
		t := make([]reflect.Type, r.ColumnCount())
		n := r.ColumnNames()
		v := make([]string, r.ColumnCount())
		for i := range t {
			t[i] = typeText
		}

		// Result loop
		for {
			row, err := r.Next(t...)
			if err != nil {
				return err
			} else if row == nil {
				break
			} else if fn != nil {
				if fn(stringSliceFromInterface(row, v), n) {
					// Set abort transaction on next iteration
					r.err = SQLITE_ABORT
				}
			}
		}
	}

	// Return success
	return nil
}

func (c *ConnEx) Begin(t SQTransaction) error {
	return c.Exec("BEGIN "+string(t)+" TRANSACTION", nil)
}

func (c *ConnEx) Commit() error {
	return c.Exec("COMMIT TRANSACTION", nil)
}

func (c *ConnEx) Rollback() error {
	return c.Exec("ROLLBACK TRANSACTION", nil)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// add adds a callback to the map
func (c *callback) add(conn *ConnEx) {
	c.Lock()
	c.fn[conn.userInfo()] = conn
	c.Unlock()
}

// delete removes a callback from the map
func (c *callback) delete(conn *ConnEx) {
	c.Lock()
	delete(c.fn, conn.userInfo())
	c.Unlock()
}

// get a connection from userInfo data
func (c *callback) get(r uintptr) *ConnEx {
	c.RLock()
	defer c.RUnlock()
	return c.fn[r]
}

// return userInfo data for the connection
func (c *ConnEx) userInfo() uintptr {
	return uintptr(unsafe.Pointer(c))
}

// return a slice of strings from a slice of interfaces
// where all elements in the interface slice are strings
// use pre-allocated string slice to avoid allocations
func stringSliceFromInterface(v []interface{}, result []string) []string {
	if v == nil {
		return nil
	}
	result = result[:len(v)]
	for i, v := range v {
		if v, ok := v.(string); ok {
			result[i] = v
		}
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// CALLBACKS

//export go_busy_handler
func go_busy_handler(userInfo unsafe.Pointer, n C.int) C.int {
	if c := cb.get(uintptr(userInfo)); c != nil && c.BusyHandlerFunc != nil {
		return C.int(boolToInt(c.BusyHandlerFunc(int(n))))
	} else {
		return C.int(boolToInt(true))
	}
}

//export go_progress_handler
func go_progress_handler(userInfo unsafe.Pointer) C.int {
	if c := cb.get(uintptr(userInfo)); c != nil && c.ProgressHandlerFunc != nil {
		return C.int(boolToInt(c.ProgressHandlerFunc()))
	} else {
		return C.int(boolToInt(false))
	}
}

//export go_commit_hook
func go_commit_hook(userInfo unsafe.Pointer) C.int {
	if c := cb.get(uintptr(userInfo)); c != nil && c.CommitHookFunc != nil {
		return C.int(boolToInt(c.CommitHookFunc()))
	} else {
		return C.int(boolToInt(false))
	}
}

//export go_rollback_hook
func go_rollback_hook(userInfo unsafe.Pointer) {
	if c := cb.get(uintptr(userInfo)); c != nil && c.RollbackHookFunc != nil {
		c.RollbackHookFunc()
	}
}

//export go_update_hook
func go_update_hook(userInfo unsafe.Pointer, op C.int, db, tbl *C.char, row C.sqlite3_int64) {
	if c := cb.get(uintptr(userInfo)); c != nil && c.UpdateHookFunc != nil {
		c.UpdateHookFunc(SQAction(op), C.GoString(db), C.GoString(tbl), int64(row))
	}
}

//export go_authorizer_hook
func go_authorizer_hook(userInfo unsafe.Pointer, op C.int, a1, a2, a3, a4 *C.char) C.int {
	if c := cb.get(uintptr(userInfo)); c != nil && c.AuthorizerHookFunc != nil {
		return C.int(c.AuthorizerHookFunc(SQAction(op), [4]string{C.GoString(a1), C.GoString(a2), C.GoString(a3), C.GoString(a4)}))
	} else {
		return C.int(0)
	}
}

//export go_exec_handler
func go_exec_handler(userInfo unsafe.Pointer, nargs C.int, row, cols **C.char) C.int {
	if c := cb.get(uintptr(userInfo)); c != nil && c.ExecFunc != nil {
		return C.int(boolToInt(c.ExecFunc(go_string_slice(int(nargs), row), go_string_slice(int(nargs), cols))))
	} else {
		return C.int(0)
	}
}

// Return []string from char**
func go_string_slice(len int, arr **C.char) []string {
	result := make([]string, len)
	cStrings := (*[1 << 28]*C.char)(unsafe.Pointer(arr))[:len:len]
	for i := range result {
		result[i] = C.GoString(cStrings[i])
	}
	return result
}
