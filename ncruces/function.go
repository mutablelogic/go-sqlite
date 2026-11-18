package ncruces

import (
	"database/sql/driver"
	"fmt"
	"math"
	"reflect"
	"time"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	sqlite3 "github.com/ncruces/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// createFunction creates the function on the connection, whose arguments
// are int64, float64, string, []byte or nil
func createFunction(db *sqlite3.Conn, fn sqlite.Function) error {
	flag := sqlite3.FunctionFlag(0)
	if fn.Deterministic {
		flag |= sqlite3.DETERMINISTIC
	}
	return db.CreateFunction(fn.Name, fn.Args, flag, func(ctx sqlite3.Context, arg ...sqlite3.Value) {
		args := make([]any, len(arg))
		for i, value := range arg {
			args[i] = argValue(value)
		}
		result, err := fn.Fn(args...)
		if err == nil {
			err = setResult(ctx, result)
		}
		if err != nil {
			ctx.ResultError(fmt.Errorf("%s: %w", fn.Name, err))
		}
	})
}

// argValue returns an argument's value: an int64, a float64, a string, a
// []byte, or nil
func argValue(value sqlite3.Value) any {
	switch value.Type() {
	case sqlite3.INTEGER:
		return value.Int64()
	case sqlite3.FLOAT:
		return value.Float()
	case sqlite3.TEXT:
		return value.Text()
	case sqlite3.BLOB:
		return value.Blob(nil)
	}
	return nil
}

// setResult sets a function's result: nil, a bool, an integer, a float, a
// string, a []byte, a time.Time, a driver.Valuer, or a pointer to one, which
// is NULL when it's nil
func setResult(ctx sqlite3.Context, result any) error {
	if result == nil {
		ctx.ResultNull()
		return nil
	}
	rv := reflect.ValueOf(result)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		ctx.ResultNull()
		return nil
	}
	switch v := result.(type) {
	case driver.Valuer:
		value, err := v.Value()
		if err != nil {
			return err
		}
		if _, ok := value.(driver.Valuer); ok {
			return fmt.Errorf("%T: a value which is a driver.Valuer", v)
		}
		return setResult(ctx, value)
	case time.Time:
		ctx.ResultTime(v, sqlite3.TimeFormatDefault)
		return nil
	case []byte:
		ctx.ResultBlob(v)
		return nil
	}
	switch rv.Kind() {
	case reflect.Pointer:
		return setResult(ctx, rv.Elem().Interface())
	case reflect.Bool:
		ctx.ResultBool(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ctx.ResultInt64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n := rv.Uint()
		if n > math.MaxInt64 {
			return fmt.Errorf("%d is larger than an integer", n)
		}
		ctx.ResultInt64(int64(n))
	case reflect.Float32, reflect.Float64:
		ctx.ResultFloat(rv.Float())
	case reflect.String:
		ctx.ResultText(rv.String())
	case reflect.Slice:
		if rv.Type().Elem().Kind() != reflect.Uint8 {
			return fmt.Errorf("%T can't be a result", result)
		}
		ctx.ResultBlob(rv.Bytes())
	default:
		return fmt.Errorf("%T can't be a result", result)
	}
	return nil
}
