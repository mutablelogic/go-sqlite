package sqlite3

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

type StatusType int

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SQLITE_DBSTATUS_LOOKASIDE_USED      StatusType = C.SQLITE_DBSTATUS_LOOKASIDE_USED
	SQLITE_DBSTATUS_CACHE_USED          StatusType = C.SQLITE_DBSTATUS_CACHE_USED
	SQLITE_DBSTATUS_SCHEMA_USED         StatusType = C.SQLITE_DBSTATUS_SCHEMA_USED
	SQLITE_DBSTATUS_STMT_USED           StatusType = C.SQLITE_DBSTATUS_STMT_USED
	SQLITE_DBSTATUS_LOOKASIDE_HIT       StatusType = C.SQLITE_DBSTATUS_LOOKASIDE_HIT
	SQLITE_DBSTATUS_LOOKASIDE_MISS_SIZE StatusType = C.SQLITE_DBSTATUS_LOOKASIDE_MISS_SIZE
	SQLITE_DBSTATUS_LOOKASIDE_MISS_FULL StatusType = C.SQLITE_DBSTATUS_LOOKASIDE_MISS_FULL
	SQLITE_DBSTATUS_CACHE_HIT           StatusType = C.SQLITE_DBSTATUS_CACHE_HIT
	SQLITE_DBSTATUS_CACHE_MISS          StatusType = C.SQLITE_DBSTATUS_CACHE_MISS
	SQLITE_DBSTATUS_CACHE_WRITE         StatusType = C.SQLITE_DBSTATUS_CACHE_WRITE
	SQLITE_DBSTATUS_DEFERRED_FKS        StatusType = C.SQLITE_DBSTATUS_DEFERRED_FKS
	SQLITE_DBSTATUS_CACHE_USED_SHARED   StatusType = C.SQLITE_DBSTATUS_CACHE_USED_SHARED
	SQLITE_DBSTATUS_CACHE_SPILL         StatusType = C.SQLITE_DBSTATUS_CACHE_SPILL
	SQLITE_DBSTATUS_MIN                            = SQLITE_DBSTATUS_LOOKASIDE_USED
	SQLITE_DBSTATUS_MAX                 StatusType = C.SQLITE_DBSTATUS_MAX
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s StatusType) String() string {
	switch s {
	case SQLITE_DBSTATUS_LOOKASIDE_USED:
		return "SQLITE_DBSTATUS_LOOKASIDE_USED"
	case SQLITE_DBSTATUS_CACHE_USED:
		return "SQLITE_DBSTATUS_CACHE_USED"
	case SQLITE_DBSTATUS_SCHEMA_USED:
		return "SQLITE_DBSTATUS_SCHEMA_USED"
	case SQLITE_DBSTATUS_STMT_USED:
		return "SQLITE_DBSTATUS_STMT_USED"
	case SQLITE_DBSTATUS_LOOKASIDE_HIT:
		return "SQLITE_DBSTATUS_LOOKASIDE_HIT"
	case SQLITE_DBSTATUS_LOOKASIDE_MISS_SIZE:
		return "SQLITE_DBSTATUS_LOOKASIDE_MISS_SIZE"
	case SQLITE_DBSTATUS_LOOKASIDE_MISS_FULL:
		return "SQLITE_DBSTATUS_LOOKASIDE_MISS_FULL"
	case SQLITE_DBSTATUS_CACHE_HIT:
		return "SQLITE_DBSTATUS_CACHE_HIT"
	case SQLITE_DBSTATUS_CACHE_MISS:
		return "SQLITE_DBSTATUS_CACHE_MISS"
	case SQLITE_DBSTATUS_CACHE_WRITE:
		return "SQLITE_DBSTATUS_CACHE_WRITE"
	case SQLITE_DBSTATUS_DEFERRED_FKS:
		return "SQLITE_DBSTATUS_DEFERRED_FKS"
	case SQLITE_DBSTATUS_CACHE_USED_SHARED:
		return "SQLITE_DBSTATUS_CACHE_USED_SHARED"
	case SQLITE_DBSTATUS_CACHE_SPILL:
		return "SQLITE_DBSTATUS_CACHE_SPILL"
	default:
		return "[?? Invalid StatusType value]"
	}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func (c *Conn) GetStatus(v StatusType) (int, int, error) {
	var cur, max C.int
	if err := SQError(C.sqlite3_db_status((*C.sqlite3)(c), (C.int)(v), &cur, &max, 0)); err != SQLITE_OK {
		return 0, 0, err
	} else {
		return int(cur), int(max), nil
	}
}

func (c *Conn) ResetStatus(v StatusType) error {
	var cur, max C.int
	if err := SQError(C.sqlite3_db_status((*C.sqlite3)(c), (C.int)(v), &cur, &max, 1)); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

func GetMemoryUsed() (int64, int64) {
	return int64(C.sqlite3_memory_used()), int64(C.sqlite3_memory_highwater(0))
}

func ResetMemoryUsed() {
	C.sqlite3_memory_highwater(1)
}
