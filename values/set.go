package values

import (
	"database/sql/driver"
	"fmt"
	"math"
	"reflect"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set sets a Go value where SQLite takes it: nil is NULL, a bool is 0 or 1,
// an integer is an integer, which is no larger than an int64, a float is a
// float, a string is text, a []byte is a blob, and a time.Time is text, of
// RFC 3339, with nanoseconds. A driver.Valuer is its value, and a named type
// is of its kind, such as a string. A pointer is of its value, or NULL when
// it's nil.
func Set(s Setter, value any) error {
	if value == nil {
		return s.SetNull()
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return s.SetNull()
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
		return Set(s, value)
	case time.Time:
		return s.SetText(v.Format(time.RFC3339Nano))
	case []byte:
		return s.SetBlob(v)
	}
	switch rv.Kind() {
	case reflect.Pointer:
		return Set(s, rv.Elem().Interface())
	case reflect.Bool:
		if rv.Bool() {
			return s.SetInt64(1)
		}
		return s.SetInt64(0)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return s.SetInt64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n := rv.Uint()
		if n > math.MaxInt64 {
			return fmt.Errorf("%d is larger than an integer", n)
		}
		return s.SetInt64(int64(n))
	case reflect.Float32, reflect.Float64:
		return s.SetFloat(rv.Float())
	case reflect.String:
		return s.SetText(rv.String())
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return s.SetBlob(rv.Bytes())
		}
	}
	return fmt.Errorf("%T can't be a value", value)
}
