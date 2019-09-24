/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Language component to build statements
type Language interface {
	gopi.Driver

	NewCreateTable(string, ...Column) CreateTable
	NewDropTable(string) DropTable
	NewInsert(string, ...string) InsertOrReplace
	NewSelect(Source) Select

	// Return data source
	NewSource(name string) Source

	// Return expressions
	//Expr(string) Expression
	//ExprArray(...string) []Expression
}

// Source represents a simple table source (schema, name and table alias)
type Source interface {
	Statement

	Schema(string) Source
	Alias(string) Source
}

// CreateTable statement
type CreateTable interface {
	Statement

	Schema(string) CreateTable
	IfNotExists() CreateTable
	Temporary() CreateTable
	WithoutRowID() CreateTable
	Unique(...string) CreateTable
}

// DropTable statement
type DropTable interface {
	Statement

	Schema(string) DropTable
	IfExists() DropTable
}

// InsertOrReplace represents an insert, upsert or replace
type InsertOrReplace interface {
	Statement

	Schema(string) InsertOrReplace
	DefaultValues() InsertOrReplace
}

// Select statement
type Select interface {
	Statement

	Distinct() Select
	LimitOffset(uint, uint) Select
}
