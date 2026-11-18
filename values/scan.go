package values

import (
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	timeType = reflect.TypeFor[time.Time]()
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Scan scans a value which SQLite gives into a pointer: to a bool, an
// integer, a float, a string, a []byte, which is nil when it's NULL, a
// time.Time, an any, a named type of one of those, or a pointer to one,
// which is nil when it's NULL. A sql.Scanner scans Any of the value. A value
// which doesn't fit, NULL, but for an any, a []byte or a pointer, and text
// which isn't a number or a time, are errors.
func Scan(v Value, dest any) error {
	if scanner, ok := dest.(sql.Scanner); ok {
		return scanner.Scan(Any(v))
	}
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("%T isn't a pointer", dest)
	}
	return set(v, rv.Elem())
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// set sets the Go value to SQLite's
func set(v Value, dest reflect.Value) error {
	typ := v.Type()

	// A scanner, an any, a pointer, which is nil when it's NULL, a []byte,
	// which is nil when it's NULL, or a time.Time
	if scanner, ok := dest.Addr().Interface().(sql.Scanner); ok {
		return scanner.Scan(Any(v))
	}
	switch {
	case dest.Kind() == reflect.Interface && dest.NumMethod() == 0:
		if value := Any(v); value == nil {
			dest.SetZero()
		} else {
			dest.Set(reflect.ValueOf(value))
		}
		return nil
	case dest.Kind() == reflect.Pointer:
		if typ == Null {
			dest.SetZero()
			return nil
		}
		ptr := reflect.New(dest.Type().Elem())
		if err := set(v, ptr.Elem()); err != nil {
			return err
		}
		dest.Set(ptr)
		return nil
	case dest.Kind() == reflect.Slice && dest.Type().Elem().Kind() == reflect.Uint8:
		if typ == Null {
			dest.SetZero()
		} else {
			dest.Set(reflect.ValueOf(v.Blob()).Convert(dest.Type()))
		}
		return nil
	case typ == Null:
		return fmt.Errorf("NULL can't be a %s", dest.Type())
	case dest.Type() == timeType:
		t, err := Time(v)
		if err != nil {
			return err
		}
		dest.Set(reflect.ValueOf(t))
		return nil
	}

	// A bool, an integer, a float or a string
	switch dest.Kind() {
	case reflect.Bool:
		b, err := boolValue(v, typ)
		if err != nil {
			return err
		}
		dest.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := intValue(v, typ)
		if err != nil {
			return err
		}
		if dest.OverflowInt(n) {
			return fmt.Errorf("%d is too large for a %s", n, dest.Type())
		}
		dest.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := intValue(v, typ)
		if err != nil {
			return err
		}
		if n < 0 || dest.OverflowUint(uint64(n)) {
			return fmt.Errorf("%d can't be a %s", n, dest.Type())
		}
		dest.SetUint(uint64(n))
	case reflect.Float32, reflect.Float64:
		f, err := floatValue(v, typ)
		if err != nil {
			return err
		}
		if dest.OverflowFloat(f) {
			return fmt.Errorf("%g is too large for a %s", f, dest.Type())
		}
		dest.SetFloat(f)
	case reflect.String:
		dest.SetString(v.Text())
	default:
		return fmt.Errorf("%s can't be scanned into", dest.Type())
	}
	return nil
}

// boolValue returns the value's bool, of a number, which is true when it
// isn't zero, or text which is one, such as true or 1
func boolValue(v Value, typ Type) (bool, error) {
	switch typ {
	case Integer:
		return v.Int64() != 0, nil
	case Float:
		return v.Float() != 0, nil
	case Text:
		return strconv.ParseBool(strings.TrimSpace(v.Text()))
	}
	return false, fmt.Errorf("a %s isn't a bool", typ)
}

// intValue returns the value's integer, of an integer, a whole float, or
// text which is one
func intValue(v Value, typ Type) (int64, error) {
	switch typ {
	case Integer:
		return v.Int64(), nil
	case Float:
		f := v.Float()
		if f != math.Trunc(f) || f < math.MinInt64 || f >= math.MaxInt64 {
			return 0, fmt.Errorf("%g isn't an integer", f)
		}
		return int64(f), nil
	case Text:
		return strconv.ParseInt(strings.TrimSpace(v.Text()), 10, 64)
	}
	return 0, fmt.Errorf("a %s isn't an integer", typ)
}

// floatValue returns the value's float, of a number, or text which is one
func floatValue(v Value, typ Type) (float64, error) {
	switch typ {
	case Integer:
		return float64(v.Int64()), nil
	case Float:
		return v.Float(), nil
	case Text:
		return strconv.ParseFloat(strings.TrimSpace(v.Text()), 64)
	}
	return 0, fmt.Errorf("a %s isn't a number", typ)
}
