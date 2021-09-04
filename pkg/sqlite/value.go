package sqlite

import (
	"math"
	"reflect"
	"time"

	// Import namespaces
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	timeType = reflect.TypeOf(time.Time{})
	blobType = reflect.TypeOf([]byte{})
	intType  = reflect.TypeOf(int64(0))
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// BoundValue returns a bound value from an arbitary value,
// which needs to be a scalar (not a map or slice) or a
// *time.Time, []byte
func BoundValue(v reflect.Value) (interface{}, error) {
	if v.IsValid() == false {
		return nil, nil
	}
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return nil, nil
		} else {
			return BoundValue(v.Elem())
		}
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() > math.MaxInt64 {
			return nil, ErrBadParameter.With("uint value overflow")
		} else {
			return int64(v.Uint()), nil
		}
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.String:
		return v.String(), nil
	case reflect.Slice:
		if v.IsNil() {
			return nil, nil
		}
		if v.Type() == blobType {
			return v.Interface().([]byte), nil
		}
	case reflect.Struct:
		if v.Type() == timeType {
			value := v.Interface().(time.Time)
			if value.IsZero() {
				return nil, nil
			} else {
				return v.Interface().(time.Time), nil
			}
		}
	}
	return nil, ErrBadParameter.With("Unsupported bind type: ", reflect.TypeOf(v))
}

// Unbound value returns a value of type t from an arbitary value
// which needs to be uint, int, float, string, bool, []byte, or *time.Time
// if v is nil then zero value is assigned
func UnboundValue(v interface{}, t reflect.Type) (reflect.Value, error) {
	if v == nil {
		return reflect.Zero(t), nil
	}
	// Do simple cases first
	rv := reflect.ValueOf(v)
	if rv.CanConvert(t) {
		return rv.Convert(t), nil
	}
	// More complex cases
	switch t.Kind() {
	case reflect.Bool:
		if rv.CanConvert(intType) {
			if rv.Convert(intType).Int() == 0 {
				return reflect.ValueOf(false), nil
			} else {
				return reflect.ValueOf(true), nil
			}
		}
	}
	return rv, ErrBadParameter.Withf("Unable to convert %q to %q", rv.Kind(), t.Kind())
}
