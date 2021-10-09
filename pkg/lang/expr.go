package lang

import (
	"fmt"
	"strings"
	"time"

	// Import namespaces
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/quote"
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
func V(v interface{}) SQExpr {
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
	case SQSource, SQStatement:
		return &e{v, nil, ""}
	case SQExpr:
		return &e{v, nil, ""}
	}
	// Unsupported value
	panic(fmt.Sprintf("V unsupported type %T", v))
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func (this *e) Or(v interface{}) SQExpr {
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
	case SQSource, SQStatement:
		return &e{this.v, v, "OR"}
	}
	// Unsupported value
	panic(fmt.Sprintf("V unsupported type %T", v))
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

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func lhs(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch e := v.(type) {
	case string:
		return Quote(e)
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
			return Quote(e.Format(time.RFC3339Nano))
		}
	case SQSource:
		return fmt.Sprint(e.WithAlias(""))
	case SQStatement:
		return e.Query()
	case SQExpr:
		return e.String()
	default:
		return Quote(fmt.Sprint(v))
	}
}

func rhs(op string, v interface{}) string {
	return strings.Join([]string{op, lhs(v)}, " ")
}
