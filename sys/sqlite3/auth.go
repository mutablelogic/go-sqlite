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

type (
	SQAction C.int
	SQAuth   C.int
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// Ref: http://www.sqlite.org/c3ref/c_limit_attached.html
const (
	SQLITE_CREATE_INDEX        SQAction = C.SQLITE_CREATE_INDEX
	SQLITE_CREATE_TABLE        SQAction = C.SQLITE_CREATE_TABLE
	SQLITE_CREATE_TEMP_INDEX   SQAction = C.SQLITE_CREATE_TEMP_INDEX
	SQLITE_CREATE_TEMP_TABLE   SQAction = C.SQLITE_CREATE_TEMP_TABLE
	SQLITE_CREATE_TEMP_TRIGGER SQAction = C.SQLITE_CREATE_TEMP_TRIGGER
	SQLITE_CREATE_TEMP_VIEW    SQAction = C.SQLITE_CREATE_TEMP_VIEW
	SQLITE_CREATE_TRIGGER      SQAction = C.SQLITE_CREATE_TRIGGER
	SQLITE_CREATE_VIEW         SQAction = C.SQLITE_CREATE_VIEW
	SQLITE_DELETE              SQAction = C.SQLITE_DELETE
	SQLITE_DROP_INDEX          SQAction = C.SQLITE_DROP_INDEX
	SQLITE_DROP_TABLE          SQAction = C.SQLITE_DROP_TABLE
	SQLITE_DROP_TEMP_INDEX     SQAction = C.SQLITE_DROP_TEMP_INDEX
	SQLITE_DROP_TEMP_TABLE     SQAction = C.SQLITE_DROP_TEMP_TABLE
	SQLITE_DROP_TEMP_TRIGGER   SQAction = C.SQLITE_DROP_TEMP_TRIGGER
	SQLITE_DROP_TEMP_VIEW      SQAction = C.SQLITE_DROP_TEMP_VIEW
	SQLITE_DROP_TRIGGER        SQAction = C.SQLITE_DROP_TRIGGER
	SQLITE_DROP_VIEW           SQAction = C.SQLITE_DROP_VIEW
	SQLITE_INSERT              SQAction = C.SQLITE_INSERT
	SQLITE_PRAGMA              SQAction = C.SQLITE_PRAGMA
	SQLITE_READ                SQAction = C.SQLITE_READ
	SQLITE_SELECT              SQAction = C.SQLITE_SELECT
	SQLITE_TRANSACTION         SQAction = C.SQLITE_TRANSACTION
	SQLITE_UPDATE              SQAction = C.SQLITE_UPDATE
	SQLITE_ATTACH              SQAction = C.SQLITE_ATTACH
	SQLITE_DETACH              SQAction = C.SQLITE_DETACH
	SQLITE_ALTER_TABLE         SQAction = C.SQLITE_ALTER_TABLE
	SQLITE_REINDEX             SQAction = C.SQLITE_REINDEX
	SQLITE_ANALYZE             SQAction = C.SQLITE_ANALYZE
	SQLITE_CREATE_VTABLE       SQAction = C.SQLITE_CREATE_VTABLE
	SQLITE_DROP_VTABLE         SQAction = C.SQLITE_DROP_VTABLE
	SQLITE_FUNCTION            SQAction = C.SQLITE_FUNCTION
	SQLITE_SAVEPOINT           SQAction = C.SQLITE_SAVEPOINT
	SQLITE_COPY                SQAction = C.SQLITE_COPY
	SQLITE_RECURSIVE           SQAction = C.SQLITE_RECURSIVE
	SQLITE_ACTION_MIN                   = SQLITE_CREATE_INDEX
	SQLITE_ACTION_MAX                   = SQLITE_RECURSIVE
)

const (
	SQLITE_ALLOW  SQAuth = SQAuth(SQLITE_OK) /* Operation requested is ok */
	SQLITE_DENY   SQAuth = C.SQLITE_DENY     /* Abort the SQL statement with an error */
	SQLITE_IGNORE SQAuth = C.SQLITE_IGNORE   /* Don't allow access, but don't generate an error */
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v SQAuth) String() string {
	switch v {
	case SQLITE_ALLOW:
		return "SQLITE_ALLOW"
	case SQLITE_DENY:
		return "SQLITE_DENY"
	case SQLITE_IGNORE:
		return "SQLITE_IGNORE"
	default:
		return "[?? Invalid SQAuth value]"
	}
}

func (v SQAction) String() string {
	switch v {
	case SQLITE_CREATE_INDEX:
		return "SQLITE_CREATE_INDEX"
	case SQLITE_CREATE_TABLE:
		return "SQLITE_CREATE_TABLE"
	case SQLITE_CREATE_TEMP_INDEX:
		return "SQLITE_CREATE_TEMP_INDEX"
	case SQLITE_CREATE_TEMP_TABLE:
		return "SQLITE_CREATE_TEMP_TABLE"
	case SQLITE_CREATE_TEMP_TRIGGER:
		return "SQLITE_CREATE_TEMP_TRIGGER"
	case SQLITE_CREATE_TEMP_VIEW:
		return "SQLITE_CREATE_TEMP_VIEW"
	case SQLITE_CREATE_TRIGGER:
		return "SQLITE_CREATE_TRIGGER"
	case SQLITE_CREATE_VIEW:
		return "SQLITE_CREATE_VIEW"
	case SQLITE_DELETE:
		return "SQLITE_DELETE"
	case SQLITE_DROP_INDEX:
		return "SQLITE_DROP_INDEX"
	case SQLITE_DROP_TABLE:
		return "SQLITE_DROP_TABLE"
	case SQLITE_DROP_TEMP_INDEX:
		return "SQLITE_DROP_TEMP_INDEX"
	case SQLITE_DROP_TEMP_TABLE:
		return "SQLITE_DROP_TEMP_TABLE"
	case SQLITE_DROP_TEMP_TRIGGER:
		return "SQLITE_DROP_TEMP_TRIGGER"
	case SQLITE_DROP_TEMP_VIEW:
		return "SQLITE_DROP_TEMP_VIEW"
	case SQLITE_DROP_TRIGGER:
		return "SQLITE_DROP_TRIGGER"
	case SQLITE_DROP_VIEW:
		return "SQLITE_DROP_VIEW"
	case SQLITE_INSERT:
		return "SQLITE_INSERT"
	case SQLITE_PRAGMA:
		return "SQLITE_PRAGMA"
	case SQLITE_READ:
		return "SQLITE_READ"
	case SQLITE_SELECT:
		return "SQLITE_SELECT"
	case SQLITE_TRANSACTION:
		return "SQLITE_TRANSACTION"
	case SQLITE_UPDATE:
		return "SQLITE_UPDATE"
	case SQLITE_ATTACH:
		return "SQLITE_ATTACH"
	case SQLITE_DETACH:
		return "SQLITE_DETACH"
	case SQLITE_ALTER_TABLE:
		return "SQLITE_ALTER_TABLE"
	case SQLITE_REINDEX:
		return "SQLITE_REINDEX"
	case SQLITE_ANALYZE:
		return "SQLITE_ANALYZE"
	case SQLITE_CREATE_VTABLE:
		return "SQLITE_CREATE_VTABLE"
	case SQLITE_DROP_VTABLE:
		return "SQLITE_DROP_VTABLE"
	case SQLITE_FUNCTION:
		return "SQLITE_FUNCTION"
	case SQLITE_SAVEPOINT:
		return "SQLITE_SAVEPOINT"
	case SQLITE_COPY:
		return "SQLITE_COPY"
	case SQLITE_RECURSIVE:
		return "SQLITE_RECURSIVE"
	default:
		return "[?? Invalid SQAction value]"
	}
}
