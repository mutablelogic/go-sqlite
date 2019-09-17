/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"fmt"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Database connection
type Connection interface {
	gopi.Driver

	// Prepare statement, destroy statement
	Prepare(string) (Statement, error)
	Destroy(Statement) error

	// Execute statement (without returning the rows)
	Do(Statement, ...interface{}) (Result, error)
	DoOnce(string, ...interface{}) (Result, error)

	// Query to return the rows
	Query(Statement, ...interface{}) (Rows, error)
	QueryOnce(string, ...interface{}) (Rows, error)

	// Return sqlite information
	Version() string
	Tables() []string
}

// Statement that can be executed
type Statement interface {
	// Return the statement query
	Query() string
}

// Return rows
type Rows interface {
	// Return column names
	Columns() []Column

	// Return next row of values, or nil if no more rows
	Next() []Value
}

// Return column name and declared type
type Column interface {
	Name() string
	DeclType() string
}

// A row value, which can be a string or int
type Value interface {
	DeclType() string     // Return declared type
	IsNull() bool         // Return true if value is NULL
	String() string       // Return value as string
	Int() int64           // Return value as int
	Uint() uint64         // Return value as uint
	Bool() bool           // Return value as bool
	Float() float64       // Return value as float
	Timestamp() time.Time // Return value as timestamp
	Bytes() []byte        // Return value as blob
}

// Result of an insert
type Result struct {
	LastInsertId int64
	RowsAffected uint64
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Result) String() string {
	return fmt.Sprintf("<sqlite.Result>{ LastInsertId=%v RowsAffected=%v }", r.LastInsertId, r.RowsAffected)
}
