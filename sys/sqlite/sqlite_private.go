/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package sqlite

import (
	sql "database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
	driver "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func to_values(args []interface{}) ([]sql.Value, error) {
	v := make([]sql.Value, len(args))
	for i, arg := range args {
		// Promote uint and int to int64
		switch arg.(type) {
		case int:
			v[i] = int64(arg.(int))
		case int8:
			v[i] = int64(arg.(int8))
		case int16:
			v[i] = int64(arg.(int16))
		case int32:
			v[i] = int64(arg.(int32))
		case string, int64, time.Time, bool, nil, []byte:
			v[i] = arg
		default:
			return nil, fmt.Errorf("Unsupported bind type: %v (argument %v)", reflect.TypeOf(arg), i)
		}

	}
	return v, nil
}

func to_rows(r *driver.SQLiteRows) (sq.Rows, error) {
	if r == nil {
		return nil, gopi.ErrBadParameter
	}
	this := new(resultset)
	this.rows = r

	// Populate columns
	columns := r.Columns()
	decltypes := r.DeclTypes()
	if len(columns) == 0 || len(columns) != len(decltypes) {
		return nil, gopi.ErrBadParameter
	}
	this.columns = make([]sq.Column, len(columns))
	this.values = make([]sql.Value, len(this.columns))

	for i, name := range columns {
		decltype := decltypes[i]
		if decltype == "" {
			decltype = r.ColumnTypeDatabaseTypeName(i)
		}
		if decltype == "" {
			decltype = DEFAULT_COLUMN_TYPE
		}
		this.columns[i] = &column{name, strings.ToUpper(decltype), i}
	}

	// Success
	return this, nil
}
