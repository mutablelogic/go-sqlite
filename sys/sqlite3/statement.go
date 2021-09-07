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

type Statement C.sqlite3_stmt

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Statement) String() string {
	str := "<statement"
	if s.IsBusy() {
		str += " busy"
	}
	if s.IsExplain() {
		str += " explain"
	}
	if s.IsReadonly() {
		str += " readonly"
	}
	if num_params := s.NumParams(); num_params > 0 {
		str += fmt.Sprint(" num_params=", num_params)
		params := make([]string, num_params)
		for i := 0; i < num_params; i++ {
			params[i] = s.ParamName(i + 1)
		}
		str += fmt.Sprintf(" params=%q", params)
	}
	if data_count := s.DataCount(); data_count > 0 {
		str += fmt.Sprint(" data_count=", data_count)
	}
	if col_count := s.ColumnCount(); col_count > 0 {
		str += fmt.Sprint(" col_count=", col_count)
		cols := make([]string, col_count)
		for i := 0; i < col_count; i++ {
			cols[i] = fmt.Sprintf("%v.%v.%v %v", s.ColumnDatabaseName(i), s.ColumnTableName(i), s.ColumnName(i), s.ColumnType(i))
		}
		str += fmt.Sprintf(" cols=%q", cols)
	}
	if sql := s.SQL(); sql != "" {
		str += fmt.Sprintf(" sql=%q", sql)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return next prepared statement, or first is nil
func (c *Conn) NextStatement(s *Statement) *Statement {
	return (*Statement)(C.sqlite3_next_stmt((*C.sqlite3)(c), (*C.sqlite3_stmt)(s)))
}

// Prepare query
func (c *Conn) Prepare(query string) (*Statement, string, error) {
	var cQuery, cExtra *C.char
	var cStatement *C.sqlite3_stmt

	// Populate CStrings
	if query != "" {
		cQuery = C.CString(query)
		defer C.free(unsafe.Pointer(cQuery))
	}
	// Prepare statement
	if err := SQError(C.sqlite3_prepare_v2((*C.sqlite3)(c), cQuery, -1, &cStatement, &cExtra)); err != SQLITE_OK {
		return nil, "", err
	}
	// Return prepared statement and extra string
	return (*Statement)(cStatement), C.GoString(cExtra), nil
}

// Return connection object from statement
func (s *Statement) Conn() *Conn {
	return (*Conn)(C.sqlite3_db_handle((*C.sqlite3_stmt)(s)))
}

// Reset statement
func (s *Statement) Reset() error {
	if err := SQError(C.sqlite3_reset((*C.sqlite3_stmt)(s))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// IsBusy returns true if in middle of execution
func (s *Statement) IsBusy() bool {
	return intToBool(int(C.sqlite3_stmt_busy((*C.sqlite3_stmt)(s))))
}

// IsExplain returns true if the  statement S is an EXPLAIN statement or an EXPLAIN QUERY PLAN
func (s *Statement) IsExplain() bool {
	return intToBool(int(C.sqlite3_stmt_isexplain((*C.sqlite3_stmt)(s))))
}

// IsReadonly returns true if the statement makes no direct changes to the content of the database file.
func (s *Statement) IsReadonly() bool {
	return intToBool(int(C.sqlite3_stmt_readonly((*C.sqlite3_stmt)(s))))
}

// Finalize prepared statement
func (s *Statement) Finalize() error {
	if err := SQError(C.sqlite3_finalize((*C.sqlite3_stmt)(s))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Step statement
func (s *Statement) Step() error {
	return SQError(C.sqlite3_step((*C.sqlite3_stmt)(s)))
}

// Return number of parameters expected for a statement
func (s *Statement) NumParams() int {
	return int(C.sqlite3_bind_parameter_count((*C.sqlite3_stmt)(s)))
}

// Returns parameter name for the nth parameter, which is an empty
// string if an unnamed parameter (?) or the parameter name otherwise (:a)
func (s *Statement) ParamName(index int) string {
	var cName *C.char
	cName = C.sqlite3_bind_parameter_name((*C.sqlite3_stmt)(s), C.int(index))
	if cName == nil {
		return ""
	} else {
		return C.GoString(cName)
	}
}

// Returns parameter index for a name, or zero
func (s *Statement) ParamIndex(name string) int {
	var cName *C.char

	// Set CString
	cName = C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	// Get parameter index and return it
	return int(C.sqlite3_bind_parameter_index((*C.sqlite3_stmt)(s), cName))
}

// Returns SQL associated with a statement
func (s *Statement) SQL() string {
	return C.GoString(C.sqlite3_sql((*C.sqlite3_stmt)(s)))
}

// Returns SQL associated with a statement, expanded with bound parameters
func (s *Statement) ExpandedSQL() string {
	return C.GoString(C.sqlite3_expanded_sql((*C.sqlite3_stmt)(s)))
}
