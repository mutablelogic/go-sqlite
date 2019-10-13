/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"fmt"

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

	// Insert, replace and update structs, rollback on error
	// and return number of affected rows
	Write(Flag, ...interface{}) (uint64, error)

	// Delete structs by key or rowid, rollback on error
	// and return number of affected rows
	Delete(...interface{}) (uint64, error)

	// Count number of objects of a particular class
	Count(Class) (uint64, error)
}

type Class interface {
	// Return name of the class
	Name() string
}

type StructClass interface {
	Class
}

type Object struct {
	RowId int64
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	FLAG_INSERT Flag = (1 << iota)
	FLAG_UPDATE
	FLAG_NONE Flag = 0
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Object) String() string {
	if this.RowId == 0 {
		return fmt.Sprintf("<sqobj.Object>{ <new> }")
	} else {
		return fmt.Sprintf("<sqobj.Object>{ rowid=%v }", this.RowId)
	}
}
