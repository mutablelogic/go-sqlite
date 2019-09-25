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
	class := &sqclass{name, pkgpath, object, columns, this.conn, this.log, nil, nil, nil}
	if class.insert = this.lang.Insert(name, class.ColumnNames()...); class.insert == nil {
		return nil
	} else if class.replace = this.lang.Replace(name, class.ColumnNames()...); class.replace == nil {
		return nil
	} else if class.update = this.lang.NewUpdate(name, class.ColumnNames()...).Where(this.lang.Equals("_rowid_", this.lang.Arg())); class.update == nil {
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

func (this *sqclass) statement(flags sq.Flag) sq.Statement {
	switch {
	case flags&(sq.FLAG_INSERT|sq.FLAG_UPDATE) == sq.FLAG_INSERT:
		return this.insert
	case flags&(sq.FLAG_INSERT|sq.FLAG_UPDATE) == sq.FLAG_INSERT|sq.FLAG_UPDATE:
		return this.replace
	case flags&(sq.FLAG_INSERT|sq.FLAG_UPDATE) == sq.FLAG_UPDATE:
		return this.update
	default:
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqclass) String() string {
	return fmt.Sprintf("<sqobj.Class>{ name=%v is_object=%v }", strconv.Quote(this.name), this.object)
}
