package lang

import (
	"fmt"
	"strings"

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
func Q(v ...interface{}) sqlite.SQStatement {
	// Case of calling Q without any arguments
	if len(v) == 0 {
		return &q{"SELECT NULL"}
	}
	var result []string
	for _, v := range v {
		if v == nil {
			result = append(result, "NULL")
		} else if v_, ok := v.(string); ok {
			result = append(result, v_)
		} else if v_, ok := v.(sqlite.SQExpr); ok {
			result = append(result, fmt.Sprint(v_))
		} else {
			result = append(result, fmt.Sprint(V(v)))
		}
	}
	return &q{strings.Join(result, "")}
}

///////////////////////////////////////////////////////////////////////////////
// QUERY

func (this *q) Query() string {
	return this.v
}

func (this *q) String() string {
	return this.v
}
