package sqlite3

/*
#cgo pkg-config: sqlite3
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SQError C.int

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SQLITE_OK         SQError = C.SQLITE_OK         /* Successful result */
	SQLITE_ERROR      SQError = C.SQLITE_ERROR      /* Generic error */
	SQLITE_INTERNAL   SQError = C.SQLITE_INTERNAL   /* Internal logic error in SQLite */
	SQLITE_PERM       SQError = C.SQLITE_PERM       /* Access permission denied */
	SQLITE_ABORT      SQError = C.SQLITE_ABORT      /* Callback routine requested an abort */
	SQLITE_BUSY       SQError = C.SQLITE_BUSY       /* The database file is locked */
	SQLITE_LOCKED     SQError = C.SQLITE_LOCKED     /* A table in the database is locked */
	SQLITE_NOMEM      SQError = C.SQLITE_NOMEM      /* A malloc() failed */
	SQLITE_READONLY   SQError = C.SQLITE_READONLY   /* Attempt to write a readonly database */
	SQLITE_INTERRUPT  SQError = C.SQLITE_INTERRUPT  /* Operation terminated by sqlite3_interrupt()*/
	SQLITE_IOERR      SQError = C.SQLITE_IOERR      /* Some kind of disk I/O error occurred */
	SQLITE_CORRUPT    SQError = C.SQLITE_CORRUPT    /* The database disk image is malformed */
	SQLITE_NOTFOUND   SQError = C.SQLITE_NOTFOUND   /* Unknown opcode in sqlite3_file_control() */
	SQLITE_FULL       SQError = C.SQLITE_FULL       /* Insertion failed because database is full */
	SQLITE_CANTOPEN   SQError = C.SQLITE_CANTOPEN   /* Unable to open the database file */
	SQLITE_PROTOCOL   SQError = C.SQLITE_PROTOCOL   /* Database lock protocol error */
	SQLITE_EMPTY      SQError = C.SQLITE_EMPTY      /* Internal use only */
	SQLITE_SCHEMA     SQError = C.SQLITE_SCHEMA     /* The database schema changed */
	SQLITE_TOOBIG     SQError = C.SQLITE_TOOBIG     /* String or BLOB exceeds size limit */
	SQLITE_CONSTRAINT SQError = C.SQLITE_CONSTRAINT /* Abort due to constraint violation */
	SQLITE_MISMATCH   SQError = C.SQLITE_MISMATCH   /* Data type mismatch */
	SQLITE_MISUSE     SQError = C.SQLITE_MISUSE     /* Library used incorrectly */
	SQLITE_NOLFS      SQError = C.SQLITE_NOLFS      /* Uses OS features not supported on host */
	SQLITE_AUTH       SQError = C.SQLITE_AUTH       /* Authorization denied */
	SQLITE_FORMAT     SQError = C.SQLITE_FORMAT     /* Not used */
	SQLITE_RANGE      SQError = C.SQLITE_RANGE      /* 2nd parameter to sqlite3_bind out of range */
	SQLITE_NOTADB     SQError = C.SQLITE_NOTADB     /* File opened that is not a database file */
	SQLITE_NOTICE     SQError = C.SQLITE_NOTICE     /* Notifications from sqlite3_log() */
	SQLITE_WARNING    SQError = C.SQLITE_WARNING    /* Warnings from sqlite3_log() */
	SQLITE_ROW        SQError = C.SQLITE_ROW        /* sqlite3_step() has another row ready */
	SQLITE_DONE       SQError = C.SQLITE_DONE       /* sqlite3_step() has finished executing */
)

///////////////////////////////////////////////////////////////////////////////
// ERROR IMPLEMENTATION

func (e SQError) Error() string {
	return C.GoString(C.sqlite3_errstr(C.int(e)))
}
