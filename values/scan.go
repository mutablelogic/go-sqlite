package values

import (
	"database/sql"
	"encoding"
	"encoding/json"
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
	timeType     = reflect.TypeFor[time.Time]()
	durationType = reflect.TypeFor[time.Duration]()
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Scan scans a value which SQLite gives into a pointer, the first of these
// which it points to:
//   - a sql.Scanner, such as a uuid.UUID, scans Any of the value
//   - an any is Any of the value
//   - a pointer is nil when it's NULL, or else points to the value
//   - NULL is nil, for a slice or a map, and otherwise an error
//   - a time.Time, as Time reads it, or a time.Duration, as Duration does
//   - an array of bytes of a blob's length, such as the standard library's
//     uuid.UUID, of the blob, unless it's an encoding.BinaryUnmarshaler
//   - an encoding.TextUnmarshaler, of text, or an encoding.BinaryUnmarshaler,
//     of a blob, or whichever it is, when it isn't both
//   - a json.Unmarshaler, of its JSON
//   - a bool, an integer, a float, or a string, of named types too, which a
//     number, or text which is one, fits, without overflowing
//   - a []byte, or an array of bytes, such as a [16]byte, of its length
//   - a slice, an array, a map or a struct, of its JSON
func Scan(v Value, dest any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("%T isn't a pointer, which isn't nil", dest)
	}
	if scanner, ok := dest.(sql.Scanner); ok {
		return scanner.Scan(Any(v))
	}
	return set(v, rv.Elem())
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// set sets the Go value to SQLite's
func set(v Value, dest reflect.Value) error {
	typ := v.Type()

	// A scanner, an any, a pointer, which is nil when it's NULL, or a
	// time.Time
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
	case typ == Null:
		if dest.Kind() == reflect.Slice || dest.Kind() == reflect.Map {
			dest.SetZero()
			return nil
		}
		return fmt.Errorf("NULL can't be a %s", dest.Type())
	case dest.Type() == timeType:
		t, err := Time(v)
		if err != nil {
			return err
		}
		dest.Set(reflect.ValueOf(t))
		return nil
	case dest.Type() == durationType:
		d, err := Duration(v)
		if err != nil {
			return err
		}
		dest.Set(reflect.ValueOf(d))
		return nil
	}

	// A blob, into an array of bytes of its length, such as the standard
	// library's uuid.UUID, which only unmarshals text
	if _, ok := dest.Addr().Interface().(encoding.BinaryUnmarshaler); !ok && typ == Blob && isBytes(dest.Type(), reflect.Array) {
		return setBytes(dest, v.Blob(), typ)
	}

	// Text, binary or JSON, which unmarshals itself
	if ok, err := unmarshal(v, typ, dest.Addr().Interface()); ok {
		return err
	}

	// A bool, an integer, a float, a string, bytes, or JSON
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
	case reflect.Slice:
		if dest.Type().Elem().Kind() == reflect.Uint8 {
			return setBytes(dest, v.Blob(), typ)
		}
		return scanJSON(v, typ, dest)
	case reflect.Array:
		if dest.Type().Elem().Kind() == reflect.Uint8 {
			if typ != Blob {
				return fmt.Errorf("a %s isn't a %s", typ, dest.Type())
			}
			return setBytes(dest, v.Blob(), typ)
		}
		return scanJSON(v, typ, dest)
	case reflect.Map, reflect.Struct:
		return scanJSON(v, typ, dest)
	default:
		return fmt.Errorf("%s can't be scanned into", dest.Type())
	}
	return nil
}

// setBytes sets a slice, or an array, of the same length, of bytes, or a
// named type of byte, such as a uint8, to the data
func setBytes(dest reflect.Value, data []byte, typ Type) error {
	if dest.Kind() == reflect.Array {
		if len(data) != dest.Len() {
			return fmt.Errorf("a %s of %d bytes isn't a %s", typ, len(data), dest.Type())
		}
	} else {
		dest.Set(reflect.MakeSlice(dest.Type(), len(data), len(data)))
	}
	for i, b := range data {
		dest.Index(i).SetUint(uint64(b))
	}
	return nil
}

// isBytes returns true when the type is a slice, or an array, of bytes
func isBytes(t reflect.Type, kind reflect.Kind) bool {
	return t.Kind() == kind && t.Elem().Kind() == reflect.Uint8
}

// unmarshal unmarshals the value into a pointer to a type which unmarshals
// itself: text into an encoding.TextUnmarshaler, a blob into an
// encoding.BinaryUnmarshaler, or either into whichever it is, when it isn't
// both, or JSON into a json.Unmarshaler, and returns true, or false when it
// doesn't unmarshal itself
func unmarshal(v Value, typ Type, ptr any) (bool, error) {
	text, isText := ptr.(encoding.TextUnmarshaler)
	binary, isBinary := ptr.(encoding.BinaryUnmarshaler)
	switch {
	case isText && (typ != Blob || !isBinary):
		return true, text.UnmarshalText(v.Blob())
	case isBinary:
		return true, binary.UnmarshalBinary(v.Blob())
	}
	if unmarshaler, ok := ptr.(json.Unmarshaler); ok {
		return true, unmarshaler.UnmarshalJSON(v.Blob())
	}
	return false, nil
}

// scanJSON unmarshals the value, which is JSON text, into a slice, an array,
// a map or a struct
func scanJSON(v Value, typ Type, dest reflect.Value) error {
	if typ != Text {
		return fmt.Errorf("a %s isn't JSON, for a %s", typ, dest.Type())
	}
	// It's of the value alone, and not merged into what it was, as a map's
	// keys, or a struct's fields, would be
	dest.SetZero()
	return json.Unmarshal([]byte(v.Text()), dest.Addr().Interface())
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
