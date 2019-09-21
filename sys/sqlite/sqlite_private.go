/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
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

func to_values(args []interface{}, num_input int) ([]sql.Value, error) {
	// Check incoming parameters if num_input is greater or equal to zero
	if num_input >= 0 && len(args) != num_input {
		return nil, fmt.Errorf("Expected %v bound query parameters, got %v", num_input, len(args))
	}
	// Make the array of bound values
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
	// Return the array
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
	defaulttype := sq.SupportedTypes()[0]
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
			decltype = defaulttype
		}
		nullable, _ := r.ColumnTypeNullable(i)
		this.columns[i] = &column{name, strings.ToUpper(decltype), nullable, false}
	}

	// Success
	return this, nil
}
