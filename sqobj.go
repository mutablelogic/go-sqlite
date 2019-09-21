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

// Frameworks

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

type Objects interface {
	gopi.Driver

	// RegisterStruct registers a struct against a database table
	RegisterStruct(string, interface{}) (StructClass, error)

	// ReflectStruct returns SQL table columns from a struct
	ReflectStruct(v interface{}) ([]Column, error)
}

type Class interface {
	Name() string
}

type StructClass interface {
	Class

	// Insert a new record and return the rowid of the inserted row
	Insert(interface{}) (int64, error)
}
