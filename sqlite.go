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

	// Perform operations within a transaction, rollback on
	// error
	Tx(func(Connection) error) error

	// Return sqlite information
	Version() string
	Tables() []string

	// Return statements
	NewColumn(string, string, bool) Column
	NewCreateTable(string, ...Column) CreateTable
	NewDropTable(string) DropTable
	NewInsert(string, ...string) InsertOrReplace

	// Reflect columns from struct
	Reflect(interface{}) ([]Column, error)
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
	Nullable() bool
	Query() string
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
// STATEMENTS

// CreateTable statement
type CreateTable interface {
	Statement

	Schema(string) CreateTable
	IfNotExists() CreateTable
	Temporary() CreateTable
	WithoutRowID() CreateTable
	PrimaryKey(...string) CreateTable
	Unique(...string) CreateTable
}

// DropTable statement
type DropTable interface {
	Statement

	Schema(string) DropTable
	IfExists() DropTable
}

// Insert statement
type InsertOrReplace interface {
	Statement

	Schema(string) InsertOrReplace
	DefaultValues() InsertOrReplace
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Result) String() string {
	return fmt.Sprintf("<sqlite.Result>{ LastInsertId=%v RowsAffected=%v }", r.LastInsertId, r.RowsAffected)
}

////////////////////////////////////////////////////////////////////////////////
// GRAVEYARD

/*
type Client interface {
	gopi.Driver

	// Reflect on data structure of a variable to return the rows we expect
	Reflect(v interface{}) ([]Column, error)
	PrimaryKey([]Column) (Key, error)
	//Unique([]Column) ([]Key, error)
	//Index([]Column) ([]Key, error)
}

type Column interface {
	Name() string
	Identifier() string // Either the name or custom identifier
	Type() Type
	Flag(Flag) bool
	Value(Flag) string
}

// These are various flags we use to modify when
// a table is created
const (
	FLAG_NONE     Flag = 0
	FLAG_NOT_NULL Flag = (1 << iota)
	FLAG_PRIMARY_KEY
	FLAG_UNIQUE_KEY
	FLAG_INDEX_KEY
	FLAG_NAME
	FLAG_TYPE
	FLAG_MAX = FLAG_TYPE
)
*/
