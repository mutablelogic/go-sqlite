package values

import (
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set sets a Go value where SQLite takes it, the first of these which it is:
//   - nil, or a nil pointer, is NULL
//   - a driver.Valuer is its value, such as a uuid.UUID's text
//   - a time.Time is text, of RFC 3339, with nanoseconds, or NULL when it's
//     zero
//   - a time.Duration is an integer, of nanoseconds
//   - a []byte is a blob
//   - an encoding.TextMarshaler is its text, such as a netip.Addr's, or the
//     standard library's uuid.UUID's, rather than its binary, when it's
//     both, so it's compared and read as text
//   - an encoding.BinaryMarshaler is a blob
//   - a json.Marshaler is its JSON, as text
//   - a bool is 0 or 1, an integer, which is no larger than an int64, is an
//     integer, a float is a float, and a string is text, of named types too
//   - a pointer is of its value
//   - an array of bytes, such as a [16]byte, is a blob
//   - a slice, an array, a map or a struct is its JSON, as text, such as
//     ["a","b"], which SQLite's json functions read, such as json_each
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
		if v.IsZero() {
			return s.SetNull()
		}
		return s.SetText(v.Format(time.RFC3339Nano))
	case time.Duration:
		return s.SetInt64(int64(v))
	case []byte:
		return s.SetBlob(v)
	case encoding.TextMarshaler:
		text, err := v.MarshalText()
		if err != nil {
			return err
		}
		return s.SetText(string(text))
	case encoding.BinaryMarshaler:
		data, err := v.MarshalBinary()
		if err != nil {
			return err
		}
		return s.SetBlob(data)
	case json.Marshaler:
		return setJSON(s, v)
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
			return s.SetBlob(bytesOf(rv))
		}
		return setJSON(s, value)
	case reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return s.SetBlob(bytesOf(rv))
		}
		return setJSON(s, value)
	case reflect.Map, reflect.Struct:
		return setJSON(s, value)
	}
	return fmt.Errorf("%T can't be a value", value)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// bytesOf returns the bytes of a slice, or an array, of bytes, or a named
// type of byte, such as a uint8
func bytesOf(rv reflect.Value) []byte {
	data := make([]byte, rv.Len())
	for i := range data {
		data[i] = byte(rv.Index(i).Uint())
	}
	return data
}

// setJSON sets the value's JSON, as text
func setJSON(s Setter, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.SetText(string(data))
}
