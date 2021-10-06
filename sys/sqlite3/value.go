package sqlite3

import (
	"fmt"
	"strconv"
	"unsafe"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo CFLAGS: -I../../c
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Type  int
	Value C.sqlite3_value
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SQLITE_INTEGER Type = C.SQLITE_INTEGER
	SQLITE_FLOAT   Type = C.SQLITE_FLOAT
	SQLITE_TEXT    Type = C.SQLITE_TEXT
	SQLITE_BLOB    Type = C.SQLITE_BLOB
	SQLITE_NULL    Type = C.SQLITE_NULL
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Type) String() string {
	switch t {
	case SQLITE_INTEGER:
		return "SQLITE_INTEGER"
	case SQLITE_FLOAT:
		return "SQLITE_FLOAT"
	case SQLITE_TEXT:
		return "SQLITE_TEXT"
	case SQLITE_BLOB:
		return "SQLITE_BLOB"
	case SQLITE_NULL:
		return "SQLITE_NULL"
	default:
		return "[?? Invalid Type value]"
	}
}

func (v *Value) String() string {
	if t := v.Type(); t == SQLITE_TEXT {
		return fmt.Sprint("<SQLITE_TEXT ", strconv.Quote(v.Text()), ">")
	} else if t == SQLITE_BLOB {
		return fmt.Sprint("<SQLITE_BLOB ", v.Bytes(), " bytes>")
	} else if t == SQLITE_NULL {
		return fmt.Sprint("<SQLITE_NULL>")
	} else {
		return fmt.Sprint("<", v.Type(), " ", v.Interface(), ">")
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Copy duplicates a value which needs to be released by the caller with Free
func (v *Value) Copy() *Value {
	return (*Value)(C.sqlite3_value_dup((*C.sqlite3_value)(v)))
}

// Free a duplicate value
func (v *Value) Free() {
	C.sqlite3_value_free((*C.sqlite3_value)(v))
}

// Type returns value type
func (v *Value) Type() Type {
	return Type(C.sqlite3_value_type((*C.sqlite3_value)(v)))
}

// Bytes returns value length in bytes
func (v *Value) Bytes() int {
	return int(C.sqlite3_value_bytes((*C.sqlite3_value)(v)))
}

// NoChange returns true if the value is unchanged in an UPDATE against a virtual table.
func (v *Value) NoChange() bool {
	return intToBool(int(C.sqlite3_value_nochange((*C.sqlite3_value)(v))))
}

// FromBind returns true if value originated from a bound parameter
/*
func (v *Value) FromBind() bool {
	return intToBool(int(C.sqlite3_value_frombind((*C.sqlite3_value)(v))))
}
*/

// Interface returns a go value from a sqlite value
func (v *Value) Interface() interface{} {
	switch v.Type() {
	case SQLITE_NULL:
		return nil
	case SQLITE_INTEGER:
		return v.Int64()
	case SQLITE_FLOAT:
		return v.Double()
	case SQLITE_TEXT:
		return v.Text()
	case SQLITE_BLOB:
		return v.Blob()
	default:
		panic("Unexpected value type")
	}
}

func (v *Value) Int32() int32 {
	return int32(C.sqlite3_value_int((*C.sqlite3_value)(v)))
}

func (v *Value) Int64() int64 {
	return int64(C.sqlite3_value_int64((*C.sqlite3_value)(v)))
}

func (v *Value) Double() float64 {
	return float64(C.sqlite3_value_double((*C.sqlite3_value)(v)))
}

func (v *Value) Text() string {
	len := C.sqlite3_value_bytes((*C.sqlite3_value)(v))
	if len == 0 {
		return ""
	} else {
		ptr := C.sqlite3_value_text((*C.sqlite3_value)(v))
		return C.GoStringN((*C.char)(unsafe.Pointer(ptr)), len)
	}
}

func (v *Value) Blob() []byte {
	len := C.sqlite3_value_bytes((*C.sqlite3_value)(v))
	if len == 0 {
		return []byte{}
	} else {
		return C.GoBytes(C.sqlite3_value_blob((*C.sqlite3_value)(v)), len)
	}
}
