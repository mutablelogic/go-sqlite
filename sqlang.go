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

	// Create
	NewCreateTable(string, ...Column) CreateTable
	NewCreateIndex(string, string, ...string) CreateIndex

	// Drop
	DropTable(string) Drop
	DropIndex(string) Drop
	DropTrigger(string) Drop
	DropView(string) Drop

	// Insert, replace and select
	Insert(string, ...string) InsertOrReplace
	Replace(string, ...string) InsertOrReplace
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

// CreateIndex statement
type CreateIndex interface {
	Statement

	Schema(string) CreateIndex
	Unique() CreateIndex
	IfNotExists() CreateIndex
}

// Drop (table,index,trigger,view) statement
type Drop interface {
	Statement

	Schema(string) Drop
	IfExists() Drop
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
