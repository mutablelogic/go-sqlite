/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqobj

import (
	"fmt"
	"reflect"
	"strconv"

	// Frameworks

	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqobj) NewClass(name, pkgpath string, object bool, columns []sq.Column) *sqclass {
	class := &sqclass{name, pkgpath, object, columns, nil, this.conn, this.log}
	if class.insert = this.lang.NewInsert(name, class.ColumnNames()...); class.insert == nil {
		return nil
	} else {
		return class
	}
}

func (this *sqclass) Name() string {
	return this.name
}

func (this *sqclass) ColumnNames() []string {
	names := make([]string, len(this.columns))
	for i, column := range this.columns {
		names[i] = column.Name()
	}
	return names
}

func (this *sqclass) BoundArgs(v interface{}) []interface{} {
	// Dereference the pointer
	v_ := reflect.ValueOf(v)
	for v_.Kind() == reflect.Ptr {
		v_ = v_.Elem()
	}
	// If not a stuct then return
	if v_.Kind() != reflect.Struct {
		return nil
	}
	// Enumerate struct fields
	values := make([]interface{}, len(this.columns))
	for i, column := range this.columns {
		value := v_.Field(column.Index())
		values[i] = value.Interface()
	}
	return values
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqclass) String() string {
	return fmt.Sprintf("<sqobj.Class>{ name=%v is_object=%v }", strconv.Quote(this.name), this.object)
}
