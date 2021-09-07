package sqlite3

/*
#cgo pkg-config: sqlite3
#include <sqlite3.h>
#include <stdlib.h>

// TODO: See if SQLITE_TRANSIENT is appropriate in these three methods

void _sqlite3_result_text(sqlite3_context* ctx, const char *str, int len) {
	sqlite3_result_text(ctx, str, len, SQLITE_TRANSIENT);
}

void _sqlite3_result_blob(sqlite3_context* ctx, void* data, int len) {
	sqlite3_result_blob(ctx, data, len, SQLITE_TRANSIENT);
}

void _sqlite3_result_blob64(sqlite3_context* ctx, void* data, sqlite3_uint64 len) {
	sqlite3_result_blob64(ctx, data, len, SQLITE_TRANSIENT);
}

*/
import "C"
import (
	"math"
	"reflect"
	"unsafe"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Context C.sqlite3_context
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *Context) String() string {
	str := "<context"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set error as too big, indicating that a string or BLOB is too long to represent
func (ctx *Context) ErrTooBig() {
	C.sqlite3_result_error_toobig((*C.sqlite3_context)(ctx))
}

// Set error to indicate that a memory allocation failed
func (ctx *Context) ErrNoMem() {
	C.sqlite3_result_error_nomem((*C.sqlite3_context)(ctx))
}

// Set error from an SQError code
func (ctx *Context) ErrCode(e SQError) {
	C.sqlite3_result_error_code((*C.sqlite3_context)(ctx), C.int(e))
}

// Set error from a string
func (ctx *Context) Err(v string) {
	// Convert error to C string
	var cErr *C.char
	cErr = C.CString(v)
	defer C.free(unsafe.Pointer(cErr))

	// Set error
	C.sqlite3_result_error((*C.sqlite3_context)(ctx), cErr, C.int(-1))
}

// Return user data from context
func (ctx *Context) UserData() unsafe.Pointer {
	return C.sqlite3_user_data((*C.sqlite3_context)(ctx))
}

// Set result as NULL
func (ctx *Context) ResultNull() {
	C.sqlite3_result_null((*C.sqlite3_context)(ctx))
}

// Set result as a double value
func (ctx *Context) ResultDouble(v float64) {
	C.sqlite3_result_double((*C.sqlite3_context)(ctx), C.double(v))
}

// Set result as a int32 value
func (ctx *Context) ResultInt32(v int32) {
	C.sqlite3_result_int((*C.sqlite3_context)(ctx), C.int(v))
}

// Set result as a int64 value
func (ctx *Context) ResultInt64(v int64) {
	C.sqlite3_result_int64((*C.sqlite3_context)(ctx), C.sqlite3_int64(v))
}

// Set result as a text value
func (ctx *Context) ResultText(v string) {
	h := (*reflect.StringHeader)(unsafe.Pointer(&v))
	C._sqlite3_result_text((*C.sqlite3_context)(ctx), (*C.char)(unsafe.Pointer(h.Data)), C.int(h.Len))
}

// Set result as a blob
func (ctx *Context) ResultBlob(data []byte) {
	if len(data) > math.MaxInt32 {
		C._sqlite3_result_blob64((*C.sqlite3_context)(ctx), unsafe.Pointer(&data[0]), C.sqlite3_uint64(len(data)))
	} else {
		C._sqlite3_result_blob((*C.sqlite3_context)(ctx), unsafe.Pointer(&data[0]), C.int(len(data)))
	}
}

// Set result as a interface value, return any errors from casting
func (ctx *Context) ResultInterface(v interface{}) error {
	if v == nil {
		ctx.ResultNull()
		return nil
	}
	switch v := v.(type) {
	case int:
		ctx.ResultInt64(int64(v))
	case int8:
		ctx.ResultInt64(int64(v))
	case int16:
		ctx.ResultInt64(int64(v))
	case int32:
		ctx.ResultInt64(int64(v))
	case int64:
		ctx.ResultInt64(int64(v))
	case uint:
		ctx.ResultInt64(int64(v))
	case uint8:
		ctx.ResultInt64(int64(v))
	case uint16:
		ctx.ResultInt64(int64(v))
	case uint32:
		ctx.ResultInt64(int64(v))
	case uint64:
		if v > math.MaxInt64 {
			return SQLITE_RANGE
		}
		ctx.ResultInt64(int64(v))
	case float32:
		ctx.ResultDouble(float64(v))
	case float64:
		ctx.ResultDouble(float64(v))
	case bool:
		ctx.ResultInt32(int32(boolToInt(v)))
	case string:
		ctx.ResultText(v)
	case []byte:
		ctx.ResultBlob(v)
	default:
		return SQLITE_MISMATCH
	}

	// Return success
	return nil
}

// Set result as a value
func (ctx *Context) ResultValue(v *Value) {
	if v == nil {
		C.sqlite3_result_null((*C.sqlite3_context)(ctx))
	} else {
		C.sqlite3_result_value((*C.sqlite3_context)(ctx), (*C.sqlite3_value)(v))
	}
}
