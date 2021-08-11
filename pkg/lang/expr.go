package lang

import (
	"fmt"
	"time"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type e struct {
	v  interface{}
	r  []sqlite.SQExpr
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
	}
	// Unsupported value
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func (this *e) Or(r sqlite.SQExpr) sqlite.SQExpr {
	return &e{this.v, []sqlite.SQExpr{r}, "OR"}
}

///////////////////////////////////////////////////////////////////////////////
// EXPRESSION

func (this *e) String() string {
	if this == P {
		return "?"
	}
	if this.v == nil {
		return "NULL"
	}
	switch e := this.v.(type) {
	case string:
		return sqlite.Quote(e)
	case uint, int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(this.v)
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
	default:
		return sqlite.Quote(fmt.Sprint(this.v))
	}
}

func (this *e) Query() string {
	return fmt.Sprint("SELECT ", this)
}
