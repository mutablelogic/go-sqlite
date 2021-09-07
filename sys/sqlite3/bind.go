package sqlite3

/*
#cgo pkg-config: sqlite3
#include <sqlite3.h>
#include <stdlib.h>

// TODO: Check whether we should be using SQLITE_TRANSIENT or something
// more efficient?
static inline int _sqlite3_bind_text(sqlite3_stmt* stmt, int index, char* p, int n) {
	return sqlite3_bind_text(stmt, index, p, n, SQLITE_TRANSIENT);
}
static inline int _sqlite3_bind_blob(sqlite3_stmt* stmt, int index, void* p, int n) {
	return sqlite3_bind_blob(stmt, index, p, n, SQLITE_TRANSIENT);
}
static inline int _sqlite3_bind_pointer(sqlite3_stmt* stmt, int index, void* p,char* t) {
	return sqlite3_bind_pointer(stmt, index, p, t, NULL);
}
*/
import "C"

import (
	"math"
	"unsafe"
)

const (
	// sqliteNamedPrefix removes these prefixes from the named parameter
	// for matching
	sqliteNamedPrefix = "? : @ $"
)

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Bind int, uint, float, bool, string, []byte, or nil to a statement with a
// named parameter, return any errors
func (s *Statement) BindNamedInterface(name string, value interface{}) error {
	// Get index of named parameter
	if index := s.ParamIndex(name); index < 1 {
		return SQLITE_RANGE
	} else {
		return s.BindInterface(index, value)
	}
}

// Bind int, uint, float, bool, string, []byte, or nil to a statement,
// return any errors
// TODO: Also accept time.Time, maybe custom types with Marshal and Unmarshal
func (s *Statement) BindInterface(index int, value interface{}) error {
	if value == nil {
		return s.BindNull(index)
	}
	switch v := value.(type) {
	case int:
		return s.BindInt64(index, int64(v))
	case int8:
		return s.BindInt64(index, int64(v))
	case int16:
		return s.BindInt64(index, int64(v))
	case int32:
		return s.BindInt64(index, int64(v))
	case int64:
		return s.BindInt64(index, int64(v))
	case uint:
		return s.BindInt64(index, int64(v))
	case uint8:
		return s.BindInt64(index, int64(v))
	case uint16:
		return s.BindInt64(index, int64(v))
	case uint32:
		return s.BindInt64(index, int64(v))
	case uint64:
		if v > math.MaxInt64 {
			return SQLITE_RANGE
		} else {
			return s.BindInt64(index, int64(v))
		}
	case float32:
		return s.BindDouble(index, float64(v))
	case float64:
		return s.BindDouble(index, float64(v))
	case bool:
		return s.BindInt32(index, int32(boolToInt(v)))
	case string:
		return s.BindText(index, v)
	case []byte:
		return s.BindBlob(index, v)
	default:
		return SQLITE_MISMATCH
	}
}

// Bind null
func (s *Statement) BindNull(index int) error {
	if err := SQError(C.sqlite3_bind_null((*C.sqlite3_stmt)(s), C.int(index))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Bind int32
func (s *Statement) BindInt32(index int, v int32) error {
	if err := SQError(C.sqlite3_bind_int((*C.sqlite3_stmt)(s), C.int(index), C.int(v))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Bind int64
func (s *Statement) BindInt64(index int, v int64) error {
	if err := SQError(C.sqlite3_bind_int64((*C.sqlite3_stmt)(s), C.int(index), C.sqlite3_int64(v))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Bind double
func (s *Statement) BindDouble(index int, v float64) error {
	if err := SQError(C.sqlite3_bind_double((*C.sqlite3_stmt)(s), C.int(index), C.double(v))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Bind text
func (s *Statement) BindText(index int, v string) error {
	var cText *C.char

	// Set CString and length
	cText = C.CString(v)
	cTextLen := C.int(len(v))
	defer C.free(unsafe.Pointer(cText))

	// Bind
	// TODO: See how many copies of string are made
	if err := SQError(C._sqlite3_bind_text((*C.sqlite3_stmt)(s), C.int(index), cText, cTextLen)); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Bind blob
func (s *Statement) BindBlob(index int, v []byte) error {
	var p unsafe.Pointer
	if v != nil {
		p = unsafe.Pointer(&v[0])
	}
	// TODO: See how many copies of data are made
	if err := SQError(C._sqlite3_bind_blob((*C.sqlite3_stmt)(s), C.int(index), p, C.int(len(v)))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Bind zero-length-blob
func (s *Statement) BindZeroBlob(index int, len int) error {
	if err := SQError(C.sqlite3_bind_zeroblob((*C.sqlite3_stmt)(s), C.int(index), C.int(len))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Bind zero-length-blob (uint64)
func (s *Statement) BindZeroBlob64(index int, len uint64) error {
	if err := SQError(C.sqlite3_bind_zeroblob64((*C.sqlite3_stmt)(s), C.int(index), C.sqlite3_uint64(len))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Bind pointer
func (s *Statement) BindPointer(index int, p unsafe.Pointer, t string) error {
	var cType *C.char

	// Set CString
	cType = C.CString(t)
	defer C.free(unsafe.Pointer(cType))

	// Bind
	if err := SQError(C._sqlite3_bind_pointer((*C.sqlite3_stmt)(s), C.int(index), p, cType)); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Clear bindings
func (s *Statement) ClearBindings() error {
	if err := SQError(C.sqlite3_clear_bindings((*C.sqlite3_stmt)(s))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}
