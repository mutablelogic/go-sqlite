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

	// Insert, replace and update
	Insert(string, ...string) InsertOrReplace
	Replace(string, ...string) InsertOrReplace
	NewDelete(string) Delete
	NewUpdate(string, ...string) Update

	// Select
	NewSelect(Source) Select

	// Return named data source
	NewSource(name string) Source

	// Build expressions
	Null() Expression
	Arg() Expression
	Value(interface{}) Expression
	Equals(string, Expression) Expression
	NotEquals(string, Expression) Expression
	And(...Expression) Expression
	Or(...Expression) Expression
}

// Source represents a simple table source (schema, name and table alias)
type Source interface {
	Statement

	Schema(string) Source
	Alias(string) Source
}

// Expression represents an expression used in Select
type Expression interface {
	Query() string
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

// Delete statement
type Delete interface {
	Statement

	Schema(string) Delete
	Where(Expression) Delete
}

// InsertOrReplace represents an insert or replace
type InsertOrReplace interface {
	Statement

	Schema(string) InsertOrReplace
	DefaultValues() InsertOrReplace
}

// Update represents an update
type Update interface {
	Statement

	Schema(string) Update
	Where(Expression) Update
}

// Select statement
type Select interface {
	Statement

	Distinct() Select
	Where(...Expression) Select
	LimitOffset(uint, uint) Select
}
