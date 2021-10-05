package sqlite3

/*
#cgo CFLAGS: -I../../c
#cgo LDFLAGS: -L../../c -lsqlite3
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Blob C.sqlite3_blob

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (b *Blob) String() string {
	str := "<blob"
	str += fmt.Sprint(" len=", b.Bytes())
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// OpenBlob handle with specified schema, table, column and rowid. If called with flag
// SQLITE_OPEN_READWRITE then the blob handle is opened for read/write access, otherwise
// for read-only access.
func (c *Conn) OpenBlob(schema, table, column string, rowid int64, flags OpenFlags) (*Blob, error) {
	if schema == "" {
		schema = DefaultSchema
	}

	// Set cString
	var cSchema, cColumn, cTable *C.char
	cSchema = C.CString(schema)
	cTable = C.CString(table)
	cColumn = C.CString(column)
	defer C.free(unsafe.Pointer(cColumn))
	defer C.free(unsafe.Pointer(cTable))
	defer C.free(unsafe.Pointer(cSchema))

	// Set flags - only SQLITE_OPEN_READWRITE is used
	if flags&OpenFlags(SQLITE_OPEN_READWRITE) != 0 {
		flags = OpenFlags(1)
	} else {
		flags = OpenFlags(0)
	}

	// Open block
	var b *C.sqlite3_blob
	if err := SQError(C.sqlite3_blob_open((*C.sqlite3)(c), cSchema, cTable, cColumn, C.sqlite3_int64(rowid), C.int(flags), &b)); err != SQLITE_OK {
		return nil, err
	} else {
		return (*Blob)(b), nil
	}
}

// Close a blob handle and release resources
func (b *Blob) Close() error {
	if err := SQError(C.sqlite3_blob_close((*C.sqlite3_blob)(b))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Reopen moves the blob handle to a new rowid
func (b *Blob) Reopen(rowid int64) error {
	if err := SQError(C.sqlite3_blob_reopen((*C.sqlite3_blob)(b), C.sqlite3_int64(rowid))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Bytes returns the size of the blob
func (b *Blob) Bytes() int {
	return int(C.sqlite3_blob_bytes((*C.sqlite3_blob)(b)))
}

// ReadAt reads data from a blob, starting at a specific byte offset within the blob
func (b *Blob) ReadAt(data []byte, offset int64) error {
	if int64(C.int(offset)) != offset {
		return SQLITE_RANGE
	}
	if err := SQError(C.sqlite3_blob_read((*C.sqlite3_blob)(b), unsafe.Pointer(&data[0]), C.int(len(data)), C.int(offset))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// WriteAt writes data into a blob, starting at a specific byte offset within the blob
func (b *Blob) WriteAt(data []byte, offset int64) error {
	if int64(C.int(offset)) != offset {
		return SQLITE_RANGE
	}
	if err := SQError(C.sqlite3_blob_write((*C.sqlite3_blob)(b), unsafe.Pointer(&data[0]), C.int(len(data)), C.int(offset))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}
