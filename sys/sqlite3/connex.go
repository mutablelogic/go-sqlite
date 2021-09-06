package sqlite3

/*
#cgo pkg-config: sqlite3
#include <sqlite3.h>
#include <stdlib.h>

extern int go_busy_handler(void* userInfo,int n);
static int _sqlite3_busy_handler(sqlite3* db,void* userInfo) {
	return sqlite3_busy_handler(db,go_busy_handler,userInfo);
}
*/
import "C"

import (
	"time"
	"unsafe"

	// Modules
	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ConnEx struct {
	*Conn
	BusyHandler
	ProgressHandler
}

type BusyHandler func(int) bool
type ProgressHandler func() bool

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultBusyTimeout = 5 * time.Second
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

	// Remove callback
	if err := c.SetBusyHandler(nil); err != nil {
		result = multierror.Append(result, err)
	}

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

// Set Busy Handler, use nil to remove the busy handler
func (c *ConnEx) SetBusyHandler(fn BusyHandler) error {
	c.BusyHandler = fn

	// Add busy handler
	if err := SQError(C._sqlite3_busy_handler((*C.sqlite3)(c.Conn), unsafe.Pointer(c))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

//export go_busy_handler
func go_busy_handler(c unsafe.Pointer, n C.int) C.int {
	if fn := (*ConnEx)(c).BusyHandler; fn != nil {
		return C.int(boolToInt(fn(int(n))))
	} else {
		return C.int(1)
	}
}
