package sqlite3

/*
#cgo pkg-config: sqlite3
#include <sqlite3.h>
#include <stdlib.h>
#include <stdio.h>

extern void go_config_logger(void* userInfo, int code, char* msg);
static inline int _sqlite3_config_logging(int enable) {
	if(enable) {
		return sqlite3_config(SQLITE_CONFIG_LOG, go_config_logger, NULL);
	} else {
		return sqlite3_config(SQLITE_CONFIG_LOG, NULL, NULL);
	}
}
*/
import "C"

import (
	"fmt"
	"unsafe"

	// Modules
	multierror "github.com/hashicorp/go-multierror"

	// Import into namespace
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type OpenFlags C.int
type Conn C.sqlite3

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SQLITE_OPEN_READONLY     OpenFlags = C.SQLITE_OPEN_READONLY                         // The database is opened in read-only mode. If the database does not already exist, an error is returned.
	SQLITE_OPEN_READWRITE    OpenFlags = C.SQLITE_OPEN_READWRITE                        // The database is opened for reading and writing if possible, or reading only if the file is write protected by the operating system. In either case the database must already exist, otherwise an error is returned.
	SQLITE_OPEN_CREATE       OpenFlags = C.SQLITE_OPEN_CREATE | C.SQLITE_OPEN_READWRITE // The database is created if it does not already exist
	SQLITE_OPEN_URI          OpenFlags = C.SQLITE_OPEN_URI                              // The filename can be interpreted as a URI if this flag is set.
	SQLITE_OPEN_MEMORY       OpenFlags = C.SQLITE_OPEN_MEMORY                           // The database will be opened as an in-memory database. The database is named by the "filename" argument for the purposes of cache-sharing, if shared cache mode is enabled, but the "filename" is otherwise ignored.
	SQLITE_OPEN_NOMUTEX      OpenFlags = C.SQLITE_OPEN_NOMUTEX                          // The new database connection will use the "multi-thread" threading mode. This means that separate threads are allowed to use SQLite at the same time, as long as each thread is using a different database connection.
	SQLITE_OPEN_FULLMUTEX    OpenFlags = C.SQLITE_OPEN_FULLMUTEX                        // The new database connection will use the "serialized" threading mode. This means the multiple threads can safely attempt to use the same database connection at the same time. (Mutexes will block any actual concurrency, but in this mode there is no harm in trying.)
	SQLITE_OPEN_SHAREDCACHE  OpenFlags = C.SQLITE_OPEN_SHAREDCACHE                      // The database is opened shared cache enabled, overriding the default shared cache setting provided by sqlite3_enable_shared_cache().
	SQLITE_OPEN_PRIVATECACHE OpenFlags = C.SQLITE_OPEN_PRIVATECACHE                     // The database is opened shared cache disabled, overriding the default shared cache setting provided by sqlite3_enable_shared_cache().
	SQLITE_OPEN_NOFOLLOW     OpenFlags = C.SQLITE_OPEN_NOFOLLOW                         // The database filename is not allowed to be a symbolic link
)

const (
	defaultSchema = "main"
	defaultMemory = ":memory:"
	defaultFlags  = SQLITE_OPEN_CREATE
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c *Conn) String() string {
	str := "<conn"
	if filename := c.Filename(""); filename != "" {
		str += fmt.Sprintf(" filename=%q", filename)
	}
	if readonly := c.Readonly(""); readonly {
		str += " readonly"
	}
	if autocommit := c.Autocommit(); autocommit {
		str += " autocommit"
	}
	if rowid := c.LastInsertId(); rowid != 0 {
		str += fmt.Sprint(" last_insert_id=", rowid)
	}
	if changes := c.Changes(); changes != 0 {
		str += fmt.Sprint(" rows_affected=", changes)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Open URL
func OpenUrl(url string, flags OpenFlags, vfs string) (*Conn, error) {
	return OpenPath(url, flags|SQLITE_OPEN_URI, vfs)
}

// Open Path
func OpenPath(path string, flags OpenFlags, vfs string) (*Conn, error) {
	var cVfs, cName *C.char
	var c *C.sqlite3

	initFn.Do(func() {
		C._sqlite3_config_logging(1)
	})

	// Check for thread safety
	if C.sqlite3_threadsafe() == 0 {
		return nil, ErrInternalAppError.With("sqlite library was not compiled for thread-safe operation")
	}

	// Set memory database if empty string
	if path == "" || path == defaultMemory {
		path = defaultMemory
		flags |= SQLITE_OPEN_MEMORY
	}

	// Set flags
	if flags == 0 {
		flags = defaultFlags
	}

	// Populate CStrings
	if vfs != "" {
		cVfs = C.CString(vfs)
		defer C.free(unsafe.Pointer(cVfs))
	}
	cName = C.CString(path)
	defer C.free(unsafe.Pointer(cName))

	// Call sqlite3_open_v2
	if err := SQError(C.sqlite3_open_v2(cName, &c, C.int(flags), cVfs)); err != SQLITE_OK {
		if c != nil {
			C.sqlite3_close_v2(c)
		}
		return nil, err
	}

	// Set extended error codes
	if err := SQError(C.sqlite3_extended_result_codes(c, 1)); err != SQLITE_OK {
		C.sqlite3_close_v2(c)
		return nil, err
	}

	return (*Conn)(c), nil
}

// Close Connection
func (c *Conn) Close() error {
	var result error

	// Close any active statements
	var s *Statement
	for {
		s = c.NextStatement(s)
		if s == nil {
			break
		}
		if err := s.Finalize(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Close database connection
	if err := SQError(C.sqlite3_close_v2((*C.sqlite3)(c))); err != SQLITE_OK {
		result = multierror.Append(result, err)
	}

	// Return any errors
	return result
}

// Get Filename
func (c *Conn) Filename(schema string) string {
	var cSchema *C.char

	// Set schema to default if empty string
	if schema == "" {
		schema = defaultSchema
	}

	// Populate CStrings
	cSchema = C.CString(schema)
	defer C.free(unsafe.Pointer(cSchema))

	// Call and return
	cFilename := C.sqlite3_db_filename((*C.sqlite3)(c), cSchema)
	if cFilename == nil {
		return ""
	} else {
		return C.GoString(cFilename)
	}
}

// Get Read-only state. Also returns false if database not found
func (c *Conn) Readonly(schema string) bool {
	var cSchema *C.char

	// Set schema to default if empty string
	if schema == "" {
		schema = defaultSchema
	}

	// Populate CStrings
	cSchema = C.CString(schema)
	defer C.free(unsafe.Pointer(cSchema))

	// Call and return
	r := int(C.sqlite3_db_readonly((*C.sqlite3)(c), cSchema))
	if r == -1 {
		return false
	} else {
		return intToBool(r)
	}
}

// Set extended result codes
func (c *Conn) SetExtendedResultCodes(v bool) error {
	if err := SQError(C.sqlite3_extended_result_codes((*C.sqlite3)(c), C.int(boolToInt(v)))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Cache Flush
func (c *Conn) CacheFlush() error {
	if err := SQError(C.sqlite3_db_cacheflush((*C.sqlite3)(c))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Release Memory
func (c *Conn) ReleaseMemory() error {
	if err := SQError(C.sqlite3_db_release_memory((*C.sqlite3)(c))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Return autocommit state
func (c *Conn) Autocommit() bool {
	return intToBool(int(C.sqlite3_get_autocommit((*C.sqlite3)(c))))
}

// Get last insert id
func (c *Conn) LastInsertId() int64 {
	return int64(C.sqlite3_last_insert_rowid((*C.sqlite3)(c)))
}

// Set last insert id
func (c *Conn) SetLastInsertId(v int64) {
	C.sqlite3_set_last_insert_rowid((*C.sqlite3)(c), C.sqlite3_int64(v))
}

// Get number of changes (rows affected)
func (c *Conn) Changes() int {
	return int(C.sqlite3_changes((*C.sqlite3)(c)))
}

// Interrupt all queries for connection
func (c *Conn) Interrupt() {
	C.sqlite3_interrupt((*C.sqlite3)(c))
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

//export go_config_logger
func go_config_logger(p unsafe.Pointer, code C.int, message *C.char) {
	fmt.Println(SQError(code), C.GoString(message))
}
