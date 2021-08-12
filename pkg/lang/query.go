package lang

import (
	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type q struct {
	v string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Q Creates a query or expression
func Q(v interface{}) sqlite.SQStatement {
	switch v := v.(type) {
	case string:
		if v == "" {
			return &q{"SELECT NULL"}
		} else {
			return &q{v}
		}
	default:
		return V(v)
	}
}

///////////////////////////////////////////////////////////////////////////////
// QUERY

func (this *q) Query() string {
	return this.v
}

func (this *q) String() string {
	return this.v
}
