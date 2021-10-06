package sqlite3

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SQLimit C.int

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// Ref: http://www.sqlite.org/c3ref/c_limit_attached.html
const (
	SQLITE_LIMIT_LENGTH              SQLimit = C.SQLITE_LIMIT_LENGTH              // The maximum size of any string or BLOB or table row, in bytes.
	SQLITE_LIMIT_SQL_LENGTH          SQLimit = C.SQLITE_LIMIT_SQL_LENGTH          // The maximum length of an SQL statement, in bytes.
	SQLITE_LIMIT_COLUMN              SQLimit = C.SQLITE_LIMIT_COLUMN              // The maximum number of columns in a table definition or in the result set of a SELECT or the maximum number of columns in an index or in an ORDER BY or GROUP BY clause.
	SQLITE_LIMIT_EXPR_DEPTH          SQLimit = C.SQLITE_LIMIT_EXPR_DEPTH          // The maximum depth of the parse tree on any expression.
	SQLITE_LIMIT_COMPOUND_SELECT     SQLimit = C.SQLITE_LIMIT_COMPOUND_SELECT     // The maximum number of terms in a compound SELECT statement.
	SQLITE_LIMIT_VDBE_OP             SQLimit = C.SQLITE_LIMIT_VDBE_OP             // The maximum number of instructions in a virtual machine program used to implement an SQL statement. If sqlite3_prepare_v2() or the equivalent tries to allocate space for more than this many opcodes in a single prepared statement, an SQLITE_NOMEM error is returned.
	SQLITE_LIMIT_FUNCTION_ARG        SQLimit = C.SQLITE_LIMIT_FUNCTION_ARG        // The maximum number of arguments on a function.
	SQLITE_LIMIT_ATTACHED            SQLimit = C.SQLITE_LIMIT_ATTACHED            // The maximum number of attached databases.
	SQLITE_LIMIT_LIKE_PATTERN_LENGTH SQLimit = C.SQLITE_LIMIT_LIKE_PATTERN_LENGTH // The maximum length of the pattern argument to the LIKE or GLOB operators.
	SQLITE_LIMIT_VARIABLE_NUMBER     SQLimit = C.SQLITE_LIMIT_VARIABLE_NUMBER     // The maximum index number of any parameter in an SQL statement.
	SQLITE_LIMIT_TRIGGER_DEPTH       SQLimit = C.SQLITE_LIMIT_TRIGGER_DEPTH       // The maximum depth of recursion for triggers.
	SQLITE_LIMIT_WORKER_THREADS      SQLimit = C.SQLITE_LIMIT_WORKER_THREADS      // The maximum number of auxiliary worker threads that a single prepared statement may start.
	SQLITE_LIMIT_MIN                         = SQLITE_LIMIT_LENGTH
	SQLITE_LIMIT_MAX                         = SQLITE_LIMIT_WORKER_THREADS
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (l SQLimit) String() string {
	switch l {
	case SQLITE_LIMIT_LENGTH:
		return "SQLITE_LIMIT_LENGTH"
	case SQLITE_LIMIT_SQL_LENGTH:
		return "SQLITE_LIMIT_SQL_LENGTH"
	case SQLITE_LIMIT_COLUMN:
		return "SQLITE_LIMIT_COLUMN"
	case SQLITE_LIMIT_EXPR_DEPTH:
		return "SQLITE_LIMIT_EXPR_DEPTH"
	case SQLITE_LIMIT_COMPOUND_SELECT:
		return "SQLITE_LIMIT_COMPOUND_SELECT"
	case SQLITE_LIMIT_VDBE_OP:
		return "SQLITE_LIMIT_VDBE_OP"
	case SQLITE_LIMIT_FUNCTION_ARG:
		return "SQLITE_LIMIT_FUNCTION_ARG"
	case SQLITE_LIMIT_ATTACHED:
		return "SQLITE_LIMIT_ATTACHED"
	case SQLITE_LIMIT_LIKE_PATTERN_LENGTH:
		return "SQLITE_LIMIT_LIKE_PATTERN_LENGTH"
	case SQLITE_LIMIT_VARIABLE_NUMBER:
		return "SQLITE_LIMIT_VARIABLE_NUMBER"
	case SQLITE_LIMIT_TRIGGER_DEPTH:
		return "SQLITE_LIMIT_TRIGGER_DEPTH"
	case SQLITE_LIMIT_WORKER_THREADS:
		return "SQLITE_LIMIT_WORKER_THREADS"
	default:
		return "[?? Invalid SQLimit value]"
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// GetLimit returns the current value of a run-time limit
func (c *Conn) GetLimit(key SQLimit) int {
	return int(C.sqlite3_limit((*C.sqlite3)(c), C.int(key), C.int(-1)))
}

// SetLimit changes the value of a run-time limit
func (c *Conn) SetLimit(key SQLimit, v int) int {
	return int(C.sqlite3_limit((*C.sqlite3)(c), C.int(key), C.int(v)))
}
