/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

type Flag uint

type Objects interface {
	gopi.Driver

	// RegisterStruct registers a struct against a database table
	RegisterStruct(interface{}) (StructClass, error)

	// ReflectStruct returns SQL table columns from a struct
	ReflectStruct(v interface{}) ([]Column, error)

	// Insert structs, rollback on error
	Insert(...interface{}) ([]int64, error)
	//Insert(Flag, ...interface{}) ([]int64, error)
}

type Class interface {
	// Return name of the class
	Name() string
}

type StructClass interface {
	Class
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	FLAG_INSERT Flag = (1 << iota)
	FLAG_REPLACE
	FLAG_NONE Flag = 0
)
