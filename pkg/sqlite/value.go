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
