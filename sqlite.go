/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"errors"
	"fmt"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// ERRORS

var (
	ErrUnsupportedType = errors.New("Unsupported type")
	ErrInvalidDate     = errors.New("Invalid date")
	ErrNotFound        = errors.New("Not Found")
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Connection to a database
type Connection interface {
	gopi.Driver
	Transaction

	// Perform operations within a transaction, rollback on error
	Txn(func(Transaction) error) error

	// Return sqlite information
	Version() string
	Schemas() []string
	Tables() []string
	TablesEx(schema string, include_temporary bool) []string
	ColumnsForTable(name, schema string) ([]Column, error)

	// Attach and detach other databases, schema cannot be 'main' or 'temp'
	Attach(schema, dsn string) error
	Detach(schema string) error
}

// Transaction that can be committed/rolled back
type Transaction interface {
	// Return statement anc column
	NewStatement(string) Statement
	NewColumn(name, decltype string, nullable, primary bool) Column
	NewColumnWithIndex(name, decltype string, nullable, primary bool, index int) Column

	// Execute statement (without returning the rows)
	Do(Statement, ...interface{}) (Result, error)
	DoOnce(string, ...interface{}) (Result, error)

	// Query to return the rows
	Query(Statement, ...interface{}) (Rows, error)
	QueryOnce(string, ...interface{}) (Rows, error)
}

// Statement that can be executed
type Statement interface {
	Query() string
}

// Rows increments over returned rows from a query
type Rows interface {
	// Return column names
	Columns() []Column

	// Return next row of values, or nil if no more rows
	Next() []Value
}

// Column represents the specification for a table column
type Column interface {
	Name() string
	DeclType() string
	Nullable() bool
	PrimaryKey() bool
	Index() int
	Query() string
}

// Value represents a typed value in a table
type Value interface {
	String() string       // Return value as string
	Int() int64           // Return value as int
	Bool() bool           // Return value as bool
	Float() float64       // Return value as float
	Timestamp() time.Time // Return value as timestamp
	Bytes() []byte        // Return value as blob

	// Return the column associated with the value
	Column() Column

	// Return true if value is NULL
	IsNull() bool
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
