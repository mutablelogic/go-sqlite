package sqlite

import (
	sql "database/sql/driver"
	"errors"
	"math"
	"reflect"
	"time"

	// Modules
	multierror "github.com/hashicorp/go-multierror"

	// Import namespaces
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	funcMarshalName   = "MarshalSQ"
	funcUnmarshalName = "UnmarshalSQ"
)

var (
	timeType = reflect.TypeOf(time.Time{})
	blobType = reflect.TypeOf([]byte{})
	intType  = reflect.TypeOf(int64(0))
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// BoundValue returns a bound value from an arbitary value,
// which needs to be a scalar (not a map or slice) or a
// time.Time, []byte
func BoundValue(v reflect.Value) (interface{}, error) {
	// Where value is not valid, return NULL
	if v.IsValid() == false {
		return nil, nil
	}
	// Try Ptr, Bool, Int, Unit, Float, String, time.Time and []byte
	if v_, err := boundScalarValue(v); errors.Is(err, ErrBadParameter) {
		// Bad parameter means we should try Marshal function
		return boundCustomValue(v, err)
	} else if err != nil {
		return nil, err
	} else {
		return v_, nil
	}
}

// Unbound value returns a value of type t from an arbitary value
// which needs to be uint, int, float, string, bool, []byte, or time.Time
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
	// Check for UnmarshalSQ method on pointer type
	if proto, err := unboundCustomValue(rv, t); errors.Is(err, ErrBadParameter) {
		// No custom unmarshaller found
		return reflect.Zero(t), ErrBadParameter.Withf("Unable to convert %q to %q", rv.Kind(), t.Kind())
	} else if err != nil {
		return reflect.Zero(t), err
	} else {
		return proto, nil
	}
}

// BoundValues decodes a set of values into a form which can be accepted by
// Exec or Query, which supports int64, float64, string, bool, []byte, and time.Time
// returns any errors
func BoundValues(v []interface{}) ([]sql.Value, error) {
	var errs error
	result := make([]sql.Value, len(v))
	for i, v := range v {
		if v_, err := BoundValue(reflect.ValueOf(v)); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			result[i] = v_
		}
	}
	return result, errs
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// boundScalarValue translates from a scalar value to a bound value
// and returns ErrBadParameter if not a supported type
func boundScalarValue(v reflect.Value) (interface{}, error) {
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
		return v.Int(), nil
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
	// Return unsupported type
	return nil, ErrBadParameter.With("Unsupported bind type: ", v.Type())
}

// boundCustomValue attempts to call func (t Type) MarshalSQ() (interface{}, error)
// on type to translate value into a bound scalar value
func boundCustomValue(v reflect.Value, err error) (interface{}, error) {
	// Check for MarshalSQ function
	fn := v.MethodByName(funcMarshalName)
	if !fn.IsValid() {
		// Return existing error
		return nil, err
	}
	// Call and expect two result arguments
	if result := fn.Call(nil); len(result) == 2 {
		if err, ok := result[1].Interface().(error); ok && err != nil {
			return nil, err
		} else {
			return boundScalarValue(reflect.ValueOf(result[0].Interface()))
		}
	}
	// Return internal app error
	return nil, ErrInternalAppError.With("Invalid number of arguments")
}

// unboundedCustomValue attempts to call func (t Type) UnmarshalSQ() (interface{}, error)
func unboundCustomValue(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	var elem bool
	if t.Kind() != reflect.Ptr {
		t = reflect.New(t).Type()
		elem = true
	}
	// Lookup custom function, return ErrBadParameter if not found
	fn, exists := t.MethodByName(funcUnmarshalName)
	if !exists {
		return reflect.Zero(t), ErrBadParameter
	}
	// Create prototype object and fill it
	proto := reflect.New(t.Elem())
	result := fn.Func.Call([]reflect.Value{proto, v})
	if len(result) != 1 {
		return reflect.Zero(t), ErrInternalAppError.With("Invalid number of return arguments")
	}
	// Check for errors
	err, ok := result[0].Interface().(error)
	if ok && err != nil {
		return reflect.Zero(t), err
	}
	// Convert back into an element if needed
	if elem {
		proto = proto.Elem()
	}
	// Return object
	return proto, nil
}
