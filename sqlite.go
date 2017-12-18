/*
	SQLite client
	(c) Copyright David Thorpe 2017
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"github.com/djthorpe/gopi"
)

type Type uint
type Flag uint

type Client interface {
	gopi.Driver

	// Reflect on data structure of a variable to return the rows we expect
	Reflect(v interface{}) ([]Column, error)

	// Create table
	CreateTable(table string, columns []Column)
}

type Column interface {
	Name() string
	Type() Type
	Flags() Flag
}

// These are the types we store 'natively' in SQLite
// in reality, they are converted from the basic types
// that SQLite stores
const (
	TYPE_NONE Type = iota
	TYPE_TEXT
	TYPE_INT
	TYPE_UINT
	TYPE_BOOL
	TYPE_FLOAT
	TYPE_BLOB
	TYPE_TIME
	TYPE_DURATION
	TYPE_MAX
)

// These are various flags we use to modify when
// a table is created
const (
	FLAG_NONE     Flag = 0
	FLAG_NOT_NULL Flag = (1 << iota)
	FLAG_UNIQUE
	FLAG_PRIMARY_KEY
	FLAG_MAX
)

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Type) String() string {
	switch t {
	case TYPE_NONE:
		return "TYPE_NONE"
	case TYPE_TEXT:
		return "TYPE_TEXT"
	case TYPE_INT:
		return "TYPE_INT"
	case TYPE_UINT:
		return "TYPE_UINT"
	case TYPE_BOOL:
		return "TYPE_BOOL"
	case TYPE_FLOAT:
		return "TYPE_FLOAT"
	case TYPE_BLOB:
		return "TYPE_BLOB"
	case TYPE_TIME:
		return "TYPE_TIME"
	case TYPE_DURATION:
		return "TYPE_DURATION"
	default:
		return "[?? Invalid Type value]"
	}
}
