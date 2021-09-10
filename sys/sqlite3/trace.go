package sqlite3

/*
#cgo pkg-config: sqlite3
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"
import "strings"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TraceType uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SQLITE_TRACE_STMT    TraceType = C.SQLITE_TRACE_STMT
	SQLITE_TRACE_PROFILE TraceType = C.SQLITE_TRACE_PROFILE
	SQLITE_TRACE_ROW     TraceType = C.SQLITE_TRACE_ROW
	SQLITE_TRACE_CLOSE   TraceType = C.SQLITE_TRACE_CLOSE
	SQLITE_TRACE_MIN               = SQLITE_TRACE_STMT
	SQLITE_TRACE_MAX               = SQLITE_TRACE_CLOSE
	SQLITE_TRACE_NONE    TraceType = 0
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v TraceType) String() string {
	if v == SQLITE_TRACE_NONE {
		return v.StringFlag()
	}
	str := ""
	for f := SQLITE_TRACE_MIN; f <= SQLITE_TRACE_MAX; f = f << 1 {
		if v&f == f {
			str += "|" + f.StringFlag()
		}
	}
	return strings.TrimPrefix(str, "|")
}

func (v TraceType) StringFlag() string {
	switch v {
	case SQLITE_TRACE_NONE:
		return "SQLITE_TRACE_NONE"
	case SQLITE_TRACE_STMT:
		return "SQLITE_TRACE_STMT"
	case SQLITE_TRACE_PROFILE:
		return "SQLITE_TRACE_PROFILE"
	case SQLITE_TRACE_ROW:
		return "SQLITE_TRACE_ROW"
	case SQLITE_TRACE_CLOSE:
		return "SQLITE_TRACE_CLOSE"
	default:
		return "[?? Invalid TraceType value]"
	}
}
