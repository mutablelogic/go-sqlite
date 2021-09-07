package sqlite3

/*
#cgo pkg-config: sqlite3
#include <sqlite3.h>
#include <stdlib.h>

extern void go_func_callback(sqlite3_context*, int, sqlite3_value**);
extern void go_step_callback(sqlite3_context*, int, sqlite3_value**);
extern void go_final_callback(sqlite3_context*);
extern void go_destroy_callback(void*);

static inline int _sqlite3_create_function_v2_scalar(sqlite3 *db,const char *name,int nargs,int flags,void* userInfo) {
	return sqlite3_create_function_v2(db,name,nargs,flags,userInfo,go_func_callback,NULL,NULL,go_destroy_callback);
}

static inline int _sqlite3_create_function_v2_aggregate(sqlite3 *db,const char *name,int nargs,int flags,void* userInfo) {
	return sqlite3_create_function_v2(db,name,nargs,flags,userInfo,NULL,go_step_callback,go_final_callback,go_destroy_callback);
}
*/
import "C"

import (
	"math/rand"
	"sync"
	"unsafe"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	StepFunc  func(*Context, []*Value)
	FinalFunc func(*Context)
)

type function struct {
	Func  StepFunc
	Step  StepFunc
	Final FinalFunc
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	mapFuncLock sync.RWMutex
	mapFuncId   int
	mapFunc     = make(map[int]function)
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create a custom function
func (c *ConnEx) CreateScalarFunction(name string, nargs int, deterministic bool, fn StepFunc) error {
	// Convert name to C string
	var cName *C.char
	cName = C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	// Set deterministic
	flags := C.int(C.SQLITE_UTF8)
	if deterministic {
		flags |= C.SQLITE_DETERMINISTIC
	}

	// Set function
	userInfo := setMapFunc(function{Func: fn})

	// Call create
	if err := SQError(C._sqlite3_create_function_v2_scalar((*C.sqlite3)(c.Conn), cName, C.int(nargs), flags, unsafe.Pointer(uintptr(userInfo)))); err != SQLITE_OK {
		return err
	}

	// Return success
	return nil
}

// TODO: CreateAggregateFunction

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func setMapFunc(fn function) int {
	mapFuncLock.Lock()
	defer mapFuncLock.Unlock()
	id := nextMapFuncId()
	mapFunc[id] = fn
	return id
}

func nextMapFuncId() int {
	for {
		mapFuncId = rand.Int()
		if _, exists := mapFunc[mapFuncId]; !exists {
			return mapFuncId
		}
	}
}

func values(n int, v **C.sqlite3_value) []*Value {
	if n == 0 {
		return []*Value{}
	} else if n < 100 {
		return (*[99](*Value))(unsafe.Pointer(v))[:n:n]
	} else {
		panic("Too many function arguments")
	}
}

//export go_func_callback
func go_func_callback(ctx *C.sqlite3_context, n C.int, v **C.sqlite3_value) {
	id := int(uintptr(C.sqlite3_user_data(ctx)))

	mapFuncLock.RLock()
	fn, exists := mapFunc[id]
	mapFuncLock.RUnlock()

	if exists && fn.Func != nil {
		fn.Func((*Context)(ctx), values(int(n), v))
	}
}

//export go_step_callback
func go_step_callback(ctx *C.sqlite3_context, n C.int, v **C.sqlite3_value) {
	id := int(uintptr(C.sqlite3_user_data(ctx)))

	mapFuncLock.RLock()
	fn, exists := mapFunc[id]
	mapFuncLock.RUnlock()

	if exists && fn.Step != nil {
		fn.Step((*Context)(ctx), values(int(n), v))
	}
}

//export go_final_callback
func go_final_callback(ctx *C.sqlite3_context) {
	id := int(uintptr(C.sqlite3_user_data(ctx)))

	mapFuncLock.RLock()
	fn, exists := mapFunc[id]
	mapFuncLock.RUnlock()

	if exists && fn.Final != nil {
		fn.Final((*Context)(ctx))
	}
}

//export go_destroy_callback
func go_destroy_callback(userInfo unsafe.Pointer) {
	id := int(uintptr(userInfo))
	mapFuncLock.Lock()
	delete(mapFunc, id)
	mapFuncLock.Unlock()
}
