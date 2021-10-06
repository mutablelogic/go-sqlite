package sqlite3

import (
	"fmt"
	"reflect"
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
// METHODS

// Return count
func (s *Statement) DataCount() int {
	return int(C.sqlite3_data_count((*C.sqlite3_stmt)(s)))
}

// Return count
func (s *Statement) ColumnCount() int {
	return int(C.sqlite3_column_count((*C.sqlite3_stmt)(s)))
}

// Return column name
func (s *Statement) ColumnName(index int) string {
	return C.GoString(C.sqlite3_column_name((*C.sqlite3_stmt)(s), C.int(index)))
}

// Return length
func (s *Statement) ColumnBytes(index int) int {
	return int(C.sqlite3_column_bytes((*C.sqlite3_stmt)(s), C.int(index)))
}

// Return database name
func (s *Statement) ColumnDatabaseName(index int) string {
	return C.GoString(C.sqlite3_column_database_name((*C.sqlite3_stmt)(s), C.int(index)))
}

// Return origin name
func (s *Statement) ColumnOriginName(index int) string {
	return C.GoString(C.sqlite3_column_origin_name((*C.sqlite3_stmt)(s), C.int(index)))
}

// Return table name
func (s *Statement) ColumnTableName(index int) string {
	return C.GoString(C.sqlite3_column_table_name((*C.sqlite3_stmt)(s), C.int(index)))
}

// Return type
func (s *Statement) ColumnType(index int) Type {
	return Type(C.sqlite3_column_type((*C.sqlite3_stmt)(s), C.int(index)))
}

// Return declared type
func (s *Statement) ColumnDeclType(index int) string {
	return C.GoString(C.sqlite3_column_decltype((*C.sqlite3_stmt)(s), C.int(index)))
}

// Return int32
func (s *Statement) ColumnInt32(index int) int32 {
	return int32(C.sqlite3_column_int((*C.sqlite3_stmt)(s), C.int(index)))
}

// Return int64
func (s *Statement) ColumnInt64(index int) int64 {
	return int64(C.sqlite3_column_int64((*C.sqlite3_stmt)(s), C.int(index)))
}

// Return float64
func (s *Statement) ColumnDouble(index int) float64 {
	return float64(C.sqlite3_column_double((*C.sqlite3_stmt)(s), C.int(index)))
}

// Return string
func (s *Statement) ColumnText(index int) string {
	// TODO: This might make many copies of the data? Look into this
	if len := s.ColumnBytes(index); len == 0 {
		return ""
	} else {
		return C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text((*C.sqlite3_stmt)(s), C.int(index)))), C.int(len))
	}
}

// Return blob, copy data if clone is true
func (s *Statement) ColumnBlob(index int, clone bool) []byte {
	// Got blob contents
	p := C.sqlite3_column_blob((*C.sqlite3_stmt)(s), C.int(index))
	if p == nil {
		return nil
	}
	// Get length of blob
	len := s.ColumnBytes(index)
	if len == 0 {
		return []byte{}
	}

	// Set up slice
	var data reflect.SliceHeader
	data.Data = uintptr(p)
	data.Len = len
	data.Cap = len

	// Return slice
	if clone {
		dst := make([]byte, len)
		copy(dst, *(*[]byte)(unsafe.Pointer(&data)))
		return dst
	} else {
		return *(*[]byte)(unsafe.Pointer(&data))
	}
}

// Return column as interface
func (s *Statement) ColumnInterface(index int) interface{} {
	t := s.ColumnType(index)
	switch t {
	case SQLITE_INTEGER:
		return s.ColumnInt64(index)
	case SQLITE_FLOAT:
		return s.ColumnDouble(index)
	case SQLITE_TEXT:
		return s.ColumnText(index)
	case SQLITE_NULL:
		return nil
	case SQLITE_BLOB:
		return s.ColumnBlob(index, true)
	}
	panic(fmt.Sprint("Bad type returned for column:", t))
}
