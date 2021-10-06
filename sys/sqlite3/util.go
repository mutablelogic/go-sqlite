package sqlite3

import (
	"time"
	"unsafe"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Convert boolean to integer
func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// Convert integer to boolean
func intToBool(v int) bool {
	if v == 0 {
		return false
	}
	return true
}

// Return version
func Version() (string, int, string) {
	return C.GoString(C.sqlite3_libversion()), int(C.sqlite3_libversion_number()), C.GoString(C.sqlite3_sourceid())
}

// Determine If An SQL Statement Is Complete
func IsComplete(v string) bool {
	var cStr *C.char

	// Populate CString
	cStr = C.CString(v)
	defer C.free(unsafe.Pointer(cStr))

	// Call and return boolean
	return intToBool(int(C.sqlite3_complete(cStr)))
}

// Enable shared cache - potentially deprecated
/*
func EnableSharedCache(v bool) error {
	if err := SQError(C.sqlite3_enable_shared_cache(C.int(boolToInt(v)))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}
*/

// Return number of keywords
func KeywordCount() int {
	return int(C.sqlite3_keyword_count())
}

// Return keyword
func KeywordName(index int) string {
	var cStr *C.char
	var cLen C.int

	if err := SQError(C.sqlite3_keyword_name(C.int(index), &cStr, &cLen)); err != SQLITE_OK {
		return ""
	} else {
		return C.GoStringN(cStr, cLen)
	}
}

// Lookup keyword
func KeywordCheck(v string) bool {
	var cStr *C.char
	var cLen C.int

	// Populate CString
	cStr = C.CString(v)
	cLen = C.int(len(v))
	defer C.free(unsafe.Pointer(cStr))

	// Return check
	return intToBool(int(C.sqlite3_keyword_check(cStr, cLen)))
}

// Sleep
func Sleep(d time.Duration) {
	C.sqlite3_sleep(C.int(d / time.Millisecond))
}
