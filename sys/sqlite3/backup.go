package sqlite3

/*
#cgo pkg-config: sqlite3
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

type Backup C.sqlite3_backup

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (b *Backup) String() string {
	str := "<backup"
	if count := b.PageCount(); count > 0 {
		str += fmt.Sprint(" page_count=", count)
	}
	if rem := b.Remaining(); rem > 0 {
		str += fmt.Sprint(" remaining=", rem)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func (c *Conn) OpenBackup(dest *Conn, destSchema, sourceSchema string) (*Backup, error) {
	if destSchema == "" {
		destSchema = defaultSchema
	}
	if sourceSchema == "" {
		sourceSchema = defaultSchema
	}

	// Set CStrings
	var cDestSchema, cSourceSchema *C.char
	cDestSchema = C.CString(destSchema)
	defer C.free(unsafe.Pointer(cDestSchema))
	cSourceSchema = C.CString(sourceSchema)
	defer C.free(unsafe.Pointer(cSourceSchema))

	// Open Backup
	if b := C.sqlite3_backup_init((*C.sqlite3)(dest), cDestSchema, (*C.sqlite3)(c), cSourceSchema); b == nil {
		return nil, SQError(C.sqlite3_errcode((*C.sqlite3)(c)))
	} else {
		return (*Backup)(b), nil
	}
}

// Finish releases all resources associated with the backup process
func (b *Backup) Finish() error {
	if err := SQError(C.sqlite3_backup_finish((*C.sqlite3_backup)(b))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Remaining returns the number of pages still to be backed up at the conclusion of the most recent Step call
func (b *Backup) Remaining() int {
	return int(C.sqlite3_backup_remaining((*C.sqlite3_backup)(b)))
}

// PageCount returns the total number of pages in the source database at the conclusion of the most recent Step call
func (b *Backup) PageCount() int {
	return int(C.sqlite3_backup_pagecount((*C.sqlite3_backup)(b)))
}

// Step copies up to n pages between the source and destination databases
func (b *Backup) Step(n int) error {
	if err := SQError(C.sqlite3_backup_step((*C.sqlite3_backup)(b), C.int(n))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}
