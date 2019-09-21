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
	"github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqobj) NewClass(name string, columns []sq.Column) *sqclass {
	class := &sqclass{name, columns, nil, this.conn, this.log}
	if class.insert = this.conn.NewInsert(name); class.insert == nil {
		return nil
	} else {
		return class
	}
}

func (this *sqclass) Name() string {
	return this.name
}

func (this *sqclass) Insert(v interface{}) (int64, error) {
	var rowid int64
	err := this.conn.Tx(func(txn sq.Connection) error {
		if args := this.BoundArgs(v); args == nil {
			return gopi.ErrBadParameter
		} else if result, err := txn.Do(this.insert, args...); err != nil {
			return err
		} else {
			rowid = result.LastInsertId
			return nil
		}
	})
	return rowid, err
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
	return fmt.Sprintf("<sqobj.Class>{ name=%v }", strconv.Quote(this.name))
}
