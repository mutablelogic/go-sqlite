package sqlite3

import (
	"time"
	"unsafe"

	// Modules
	"github.com/mutablelogic/go-sqlite/sys/sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// SetTraceHook sets a function to receive executed SQL statements, with
// the time it took to execute them. The callback is provided with the
// SQL statement. If the second argument is less than zero, the callback is preparing
// a statement for execution. If the second argument is non-zero, the
// callback is invoked when the statement is completed.
func (c *Conn) SetTraceHook(fn TraceFunc) {
	if fn == nil {
		c.ConnEx.SetTraceHook(nil, 0)
	} else {
		c.ConnEx.SetTraceHook(func(t sqlite3.TraceType, a, b unsafe.Pointer) int {
			switch t {
			case sqlite3.SQLITE_TRACE_STMT:
				s := (*sqlite3.Statement)(a)
				fn(s.SQL(), -1)
			case sqlite3.SQLITE_TRACE_PROFILE:
				s := (*sqlite3.Statement)(a)
				ns := time.Duration(time.Duration(*(*int64)(b)) * time.Nanosecond)
				fn(s.ExpandedSQL(), ns)
			}
			return 0
		}, sqlite3.SQLITE_TRACE_PROFILE|sqlite3.SQLITE_TRACE_STMT)
	}
}
