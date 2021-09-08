package sqlite3

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Results struct {
	st   *Statement
	err  error
	cols []interface{}
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	typeText = reflect.TypeOf("")
	typeBlob = reflect.TypeOf([]byte{})
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r *Results) String() string {
	str := "<results"
	if r.st != nil {
		str += " " + r.st.String()
	}
	if r.err != nil && r.err != SQLITE_ROW {
		str += fmt.Sprintf(" err=%q", r.err.Error())
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Return a new results object
func results(st *Statement, err error) *Results {
	r := new(Results)
	r.st = st
	r.err = err
	r.cols = make([]interface{}, 0, st.ColumnCount())
	return r
}

// Return next row of values, or nil if there are no more rows.
// If arguments t are provided, then the values will be
// cast to the types in t if that is possible, or else an error
// will occur
func (r *Results) Next(t ...reflect.Type) ([]interface{}, error) {
	var result error

	// If no more results, return nil,nil
	if r.err == SQLITE_DONE {
		r.st.Reset()
		r.st = nil
		r.cols = nil
		return nil, nil
	}

	// Check for SQLITE_ROW result, abort result if error occurred
	if r.err != SQLITE_ROW {
		r.st.Reset()
		r.st = nil
		r.cols = nil
		return nil, r.err
	}

	// Adjust size of columns
	n := r.st.DataCount()
	r.cols = r.cols[:n]

	// Cast values into columns. If type t is defined also cast
	// value to type t
	for i := 0; i < n; i++ {
		if len(t) > i {
			if v, err := r.castvalue(i, t[i]); err != nil {
				result = multierror.Append(result, err)
			} else {
				r.cols[i] = v
			}
		} else {
			if v, err := r.value(i); err != nil {
				result = multierror.Append(result, err)
			} else {
				r.cols[i] = v
			}
		}
	}

	// Call step to next row
	r.err = r.st.Step()

	// Return result
	return r.cols, nil
}

// Return column names for the next row to be fetched
func (r *Results) ColumnNames() []string {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnName(i)
	}
	return result
}

// Return column count
func (r *Results) ColumnCount() int {
	return r.st.ColumnCount()
}

// Return column types for the next row to be fetched
func (r *Results) ColumnTypes() []Type {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]Type, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnType(i)
	}
	return result
}

// Return column decltypes for the next row to be fetched
func (r *Results) ColumnDeclTypes() []string {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnDeclType(i)
	}
	return result
}

// Return the source database schema name for the next row to be fetched
func (r *Results) ColumnDatabaseNames() []string {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnDatabaseName(i)
	}
	return result
}

// Return the source table name for the next row to be fetched
func (r *Results) ColumnTableNames() []string {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnTableName(i)
	}
	return result
}

// Return the origin for the next row to be fetched
func (r *Results) ColumnOriginNames() []string {
	if r.st == nil {
		return nil
	}
	len := r.st.ColumnCount()
	result := make([]string, len)
	for i := 0; i < len; i++ {
		result[i] = r.st.ColumnOriginName(i)
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (r *Results) value(index int) (interface{}, error) {
	switch r.st.ColumnType(index) {
	case SQLITE_INTEGER:
		return r.st.ColumnInt64(index), nil
	case SQLITE_FLOAT:
		return r.st.ColumnDouble(index), nil
	case SQLITE_TEXT:
		return r.st.ColumnText(index), nil
	case SQLITE_BLOB:
		return r.st.ColumnBlob(index), nil
	case SQLITE_NULL:
		return nil, nil
	default:
		return nil, fmt.Errorf("Unsupported column type %d", r.st.ColumnType(index))
	}
}

func (r *Results) castvalue(index int, t reflect.Type) (interface{}, error) {
	st := r.st.ColumnType(index)

	// Check for null
	if st == SQLITE_NULL {
		return nil, nil
	}

	// Do simple cases first
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rv := reflect.ValueOf(r.st.ColumnInt64(index))
		if rv.CanConvert(t) {
			return rv.Convert(t).Interface(), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rv := reflect.ValueOf(r.st.ColumnInt64(index))
		if rv.CanConvert(t) {
			return rv.Convert(t).Interface(), nil
		}
	case reflect.Bool:
		if r.st.ColumnInt64(index) == 0 {
			return false, nil
		} else {
			return true, nil
		}
	case reflect.String:
		return r.st.ColumnText(index), nil
	case reflect.Float32, reflect.Float64:
		rv := reflect.ValueOf(r.st.ColumnDouble(index))
		if rv.CanConvert(t) {
			return rv.Convert(t).Interface(), nil
		}
	}
	// Do types
	switch t {
	case typeBlob:
		if st == SQLITE_BLOB {
			return r.st.ColumnBlob(index), nil
		} else if st == SQLITE_TEXT {
			return []byte(r.st.ColumnText(index)), nil
		}
	}

	// No conversion possible
	return nil, fmt.Errorf("Cannot convert %q to %q", r.st.ColumnType(index), t)
}
