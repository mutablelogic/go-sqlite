package lang

import (
	"fmt"
	"strings"
	"time"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type e struct {
	v  interface{}
	r  interface{}
	op string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// P defines a bound parameter
	P = &e{nil, nil, ""}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// V creates a value
func V(v interface{}) sqlite.SQExpr {
	if v == nil {
		return &e{nil, nil, ""}
	}
	switch v.(type) {
	case string:
		return &e{v, nil, ""}
	case uint, int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return &e{v, nil, ""}
	case bool:
		return &e{v, nil, ""}
	case time.Time:
		return &e{v, nil, ""}
	case sqlite.SQSource, sqlite.SQStatement:
		return &e{v, nil, ""}
	}
	// Unsupported value
	panic(fmt.Sprintf("V unsupported value %q", v))
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func (this *e) Or(v interface{}) sqlite.SQExpr {
	// TODO: if this.r is not nil, then v is this
	if v == nil {
		return &e{this.v, nil, "OR"}
	}
	switch v.(type) {
	case string:
		return &e{this.v, v, "OR"}
	case uint, int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return &e{this.v, v, "OR"}
	case bool:
		return &e{this.v, v, "OR"}
	case time.Time:
		return &e{this.v, v, "OR"}
	case sqlite.SQSource, sqlite.SQStatement:
		return &e{this.v, v, "OR"}
	}
	// Unsupported value
	panic(fmt.Sprintf("V unsupported value %q", v))
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *e) String() string {
	if this == P {
		return "?"
	}
	if this.op == "" {
		return lhs(this.v)
	} else {
		return lhs(this.v) + " " + rhs(this.op, this.r)
	}
}

func (this *e) Query() string {
	return "SELECT " + fmt.Sprint(this)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func lhs(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch e := v.(type) {
	case string:
		return sqlite.Quote(e)
	case uint, int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(v)
	case bool:
		if e {
			return "TRUE"
		} else {
			return "FALSE"
		}
	case time.Time:
		if e.IsZero() {
			return "NULL"
		} else {
			return sqlite.Quote(e.Format(time.RFC3339Nano))
		}
	case sqlite.SQSource:
		return fmt.Sprint(e.WithAlias(""))
	case sqlite.SQStatement:
		return e.Query()
	default:
		return sqlite.Quote(fmt.Sprint(v))
	}
}

func rhs(op string, v interface{}) string {
	return strings.Join([]string{op, lhs(v)}, " ")
}
