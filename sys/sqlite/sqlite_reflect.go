/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"fmt"
	"reflect"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// REFLECT IMPLEMENTATION

func (this *sqlite) Reflect(v interface{}) ([]sq.Column, error) {
	this.log.Debug2("<sqlite.Reflect>{ %+t }", v)

	// Dereference the pointer
	v_ := reflect.ValueOf(v)
	for v_.Kind() == reflect.Ptr {
		v_ = v_.Elem()
	}
	// If not a stuct then return
	if v_.Kind() != reflect.Struct {
		return nil, gopi.ErrBadParameter
	}

	// Enumerate struct fields
	for i := 0; i < v_.Type().NumField(); i++ {
		fmt.Println(v_.Field(i))
	}
	// Return columns
	return nil, gopi.ErrNotImplemented
}
