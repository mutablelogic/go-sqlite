package lang

import (
	"fmt"
	"strings"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sel struct {
	source        []SQExpr
	distinct      bool
	limit, offset uint
	where         []interface{}
	to            []SQExpr
	order         []SQSource
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// S defines a select statement
func S(sources ...SQExpr) SQSelect {
	return &sel{sources, false, 0, 0, nil, nil, nil}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *sel) WithDistinct() SQSelect {
	return &sel{this.source, true, this.limit, this.offset, this.where, this.to, this.order}
}

func (this *sel) WithLimitOffset(limit, offset uint) SQSelect {
	return &sel{this.source, this.distinct, limit, offset, this.where, this.to, this.order}
}

func (this *sel) Where(v ...interface{}) SQSelect {
	if len(v) == 0 {
		// Reset where clause
		return &sel{this.source, this.distinct, this.limit, this.offset, nil, this.to, this.order}
	}
	// Where clause with an expression
	return &sel{this.source, this.distinct, this.limit, this.offset, append(this.where, v...), this.to, this.order}
}

func (this *sel) To(v ...SQExpr) SQSelect {
	if len(v) == 0 {
		// Reset to clause
		return &sel{this.source, this.distinct, this.limit, this.offset, this.where, nil, this.order}
	}
	// To clause with an expression
	return &sel{this.source, this.distinct, this.limit, this.offset, this.where, append(this.to, v...), this.order}
}

func (this *sel) Order(v ...SQSource) SQSelect {
	if len(v) == 0 {
		// Reset order clause
		return &sel{this.source, this.distinct, this.limit, this.offset, this.where, this.to, nil}
	}
	// Append order clause
	return &sel{this.source, this.distinct, this.limit, this.offset, this.where, this.to, append(this.order, v...)}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sel) String() string {
	return this.Query()
}

func (this *sel) Query() string {
	tokens := []string{"SELECT"}

	// Where there are no sources, return SELECT NULL
	if len(this.source) == 0 && len(this.to) == 0 {
		return "SELECT NULL"
	}

	// Add distinct keyword
	if this.distinct {
		tokens = append(tokens, "DISTINCT")
	}

	// To
	if len(this.to) == 0 {
		tokens = append(tokens, "*")
	} else {
		token := ""
		for i, expr := range this.to {
			if i > 0 {
				token += ","
			}
			token += expr.String()
		}
		tokens = append(tokens, token)
	}

	// Add sources using a cross join
	if len(this.source) > 0 {
		token := "FROM "
		for i, source := range this.source {
			if i > 0 {
				token += ","
			}
			token += source.String()
		}
		tokens = append(tokens, token)
	}

	// Where clause
	if len(this.where) > 0 {
		tokens = append(tokens, "WHERE")
		for i, expr := range this.where {
			if i > 0 {
				tokens = append(tokens, "AND")
			}
			tokens = append(tokens, fmt.Sprint(V(expr)))
		}
	}

	// Order clause
	if len(this.order) > 0 {
		token := "ORDER BY "
		for i, expr := range this.order {
			if i > 0 {
				token += ","
			}
			token += fmt.Sprint(expr)
		}
		tokens = append(tokens, token)
	}

	// Add offset and limit
	if this.limit == 0 && this.offset > 0 {
		tokens = append(tokens, "LIMIT -1 OFFSET", fmt.Sprint(this.offset))
	} else if this.limit > 0 && this.offset == 0 {
		tokens = append(tokens, "LIMIT", fmt.Sprint(this.limit))
	} else if this.limit > 0 && this.offset > 0 {
		tokens = append(tokens, "LIMIT", fmt.Sprint(this.offset)+","+fmt.Sprint(this.limit))
	}

	// Return the query
	return strings.Join(tokens, " ")
}
