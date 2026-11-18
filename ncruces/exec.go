package ncruces

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	sqlite3 "github.com/ncruces/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// row is a statement's row, which is scanned into values
type row struct {
	stmt *sqlite3.Stmt
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	timeType = reflect.TypeFor[time.Time]()
)

////////////////////////////////////////////////////////////////////////////////
// EXECUTE

// execute executes the query's statements, with its ${key} replaced, and its
// @key arguments bound, and scans their rows into the reader, which is
// ErrNotFound without any
func execute(db *sqlite3.Conn, bind *sqlite.Bind, query string, reader sqlite.Reader) error {
	query, vars := bind.Expand(query)
	scanned := false
	for strings.TrimSpace(query) != "" {
		stmt, tail, err := db.Prepare(query)
		if err != nil {
			return err
		}
		query = tail
		if stmt == nil {
			// A comment, or a ;
			continue
		}
		if err := step(stmt, vars, reader, &scanned); err != nil {
			return err
		}
	}
	if reader != nil && !scanned {
		return sqlite.ErrNotFound
	}
	return nil
}

// step binds the statement's arguments, and scans its rows into the reader
func step(stmt *sqlite3.Stmt, vars map[string]any, reader sqlite.Reader, scanned *bool) error {
	defer stmt.Close()
	for i := 1; i <= stmt.BindCount(); i++ {
		name := stmt.BindName(i)
		if name == "" {
			// A positional argument, which is NULL
			continue
		}
		if err := bindValue(stmt, i, vars[name[1:]]); err != nil {
			return sqlite.ErrBadParameter.Withf("%s: %w", name, err)
		}
	}
	for stmt.Step() {
		if reader == nil {
			continue
		}
		if err := reader.Scan(row{stmt}); err != nil {
			return err
		}
		*scanned = true
	}
	return stmt.Err()
}

////////////////////////////////////////////////////////////////////////////////
// BIND

// bindValue binds a value: nil, a bool, an integer, a float, a string, a
// []byte, a time.Time, a driver.Valuer, or a pointer to one, which is NULL
// when it's nil
func bindValue(stmt *sqlite3.Stmt, i int, value any) error {
	if value == nil {
		return stmt.BindNull(i)
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return stmt.BindNull(i)
	}
	switch v := value.(type) {
	case driver.Valuer:
		value, err := v.Value()
		if err != nil {
			return err
		}
		if _, ok := value.(driver.Valuer); ok {
			return fmt.Errorf("%T: a value which is a driver.Valuer", v)
		}
		return bindValue(stmt, i, value)
	case time.Time:
		return stmt.BindTime(i, v, sqlite3.TimeFormatDefault)
	case []byte:
		return stmt.BindBlob(i, v)
	}
	switch rv.Kind() {
	case reflect.Pointer:
		return bindValue(stmt, i, rv.Elem().Interface())
	case reflect.Bool:
		return stmt.BindBool(i, rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return stmt.BindInt64(i, rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if n := rv.Uint(); n > math.MaxInt64 {
			return fmt.Errorf("%d is larger than an integer", n)
		} else {
			return stmt.BindInt64(i, int64(n))
		}
	case reflect.Float32, reflect.Float64:
		return stmt.BindFloat(i, rv.Float())
	case reflect.String:
		return stmt.BindText(i, rv.String())
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return stmt.BindBlob(i, rv.Bytes())
		}
	}
	return fmt.Errorf("%T can't be bound", value)
}

////////////////////////////////////////////////////////////////////////////////
// SCAN

// Scan scans the row's columns into the values, which are pointers, of each
// of them: to a bool, an integer, a float, a string, a []byte, a time.Time,
// an any, or a pointer to one, which is nil when it's NULL, or a
// sql.Scanner.
func (r row) Scan(dest ...any) error {
	if n := r.stmt.ColumnCount(); len(dest) != n {
		return fmt.Errorf("%d values, for %d columns", len(dest), n)
	}
	for i, dest := range dest {
		if err := scan(r.stmt, i, dest); err != nil {
			return fmt.Errorf("column %q: %w", r.stmt.ColumnName(i), err)
		}
	}
	return nil
}

// scan scans a column into a value
func scan(stmt *sqlite3.Stmt, i int, dest any) error {
	if scanner, ok := dest.(sql.Scanner); ok {
		return scanner.Scan(columnValue(stmt, i))
	}
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("%T isn't a pointer", dest)
	}
	return set(stmt, i, rv.Elem())
}

// set sets the value to the column's
func set(stmt *sqlite3.Stmt, i int, v reflect.Value) error {
	typ := stmt.ColumnType(i)

	// A scanner, an any, a pointer, which is nil when it's NULL, a
	// time.Time, or a []byte, which is nil when it's NULL
	if scanner, ok := v.Addr().Interface().(sql.Scanner); ok {
		return scanner.Scan(columnValue(stmt, i))
	}
	switch {
	case v.Kind() == reflect.Interface && v.NumMethod() == 0:
		if value := columnValue(stmt, i); value == nil {
			v.SetZero()
		} else {
			v.Set(reflect.ValueOf(value))
		}
		return nil
	case v.Kind() == reflect.Pointer:
		if typ == sqlite3.NULL {
			v.SetZero()
			return nil
		}
		ptr := reflect.New(v.Type().Elem())
		if err := set(stmt, i, ptr.Elem()); err != nil {
			return err
		}
		v.Set(ptr)
		return nil
	case v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8:
		if typ == sqlite3.NULL {
			v.SetZero()
		} else {
			v.Set(reflect.ValueOf(stmt.ColumnBlob(i, nil)).Convert(v.Type()))
		}
		return nil
	case typ == sqlite3.NULL:
		return fmt.Errorf("NULL can't be a %s", v.Type())
	case v.Type() == timeType:
		t := stmt.ColumnTime(i, sqlite3.TimeFormatAuto)
		if t.IsZero() && (typ == sqlite3.TEXT || typ == sqlite3.BLOB) {
			return fmt.Errorf("%q isn't a time", stmt.ColumnText(i))
		}
		v.Set(reflect.ValueOf(t))
		return nil
	}

	// A bool, an integer, a float or a string
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(stmt.ColumnBool(i))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := intValue(stmt, i, typ)
		if err != nil {
			return err
		}
		if v.OverflowInt(n) {
			return fmt.Errorf("%d is too large for a %s", n, v.Type())
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := intValue(stmt, i, typ)
		if err != nil {
			return err
		}
		if n < 0 || v.OverflowUint(uint64(n)) {
			return fmt.Errorf("%d can't be a %s", n, v.Type())
		}
		v.SetUint(uint64(n))
	case reflect.Float32, reflect.Float64:
		f, err := floatValue(stmt, i, typ)
		if err != nil {
			return err
		}
		if v.OverflowFloat(f) {
			return fmt.Errorf("%g is too large for a %s", f, v.Type())
		}
		v.SetFloat(f)
	case reflect.String:
		v.SetString(stmt.ColumnText(i))
	default:
		return fmt.Errorf("%s can't be scanned into", v.Type())
	}
	return nil
}

// columnValue returns the column's value: an int64, a float64, a string, a
// []byte, or nil
func columnValue(stmt *sqlite3.Stmt, i int) any {
	switch stmt.ColumnType(i) {
	case sqlite3.INTEGER:
		return stmt.ColumnInt64(i)
	case sqlite3.FLOAT:
		return stmt.ColumnFloat(i)
	case sqlite3.TEXT:
		return stmt.ColumnText(i)
	case sqlite3.BLOB:
		return stmt.ColumnBlob(i, nil)
	}
	return nil
}

// intValue returns the column's integer, of an integer, a whole float, or
// text which is one
func intValue(stmt *sqlite3.Stmt, i int, typ sqlite3.Datatype) (int64, error) {
	switch typ {
	case sqlite3.INTEGER:
		return stmt.ColumnInt64(i), nil
	case sqlite3.FLOAT:
		f := stmt.ColumnFloat(i)
		if f != math.Trunc(f) || f < math.MinInt64 || f >= math.MaxInt64 {
			return 0, fmt.Errorf("%g isn't an integer", f)
		}
		return int64(f), nil
	case sqlite3.TEXT:
		return strconv.ParseInt(strings.TrimSpace(stmt.ColumnText(i)), 10, 64)
	}
	return 0, fmt.Errorf("a blob isn't an integer")
}

// floatValue returns the column's float, of a number, or text which is one
func floatValue(stmt *sqlite3.Stmt, i int, typ sqlite3.Datatype) (float64, error) {
	switch typ {
	case sqlite3.INTEGER, sqlite3.FLOAT:
		return stmt.ColumnFloat(i), nil
	case sqlite3.TEXT:
		return strconv.ParseFloat(strings.TrimSpace(stmt.ColumnText(i)), 64)
	}
	return 0, fmt.Errorf("a blob isn't a number")
}
