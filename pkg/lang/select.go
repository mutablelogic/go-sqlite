package lang

import (
	"fmt"
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sel struct {
	source        []sqlite.SQSource
	distinct      bool
	limit, offset uint
	where         []interface{}
	to            []sqlite.SQSource
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// S defines a select statement
func S(sources ...sqlite.SQSource) sqlite.SQSelect {
	return &sel{sources, false, 0, 0, nil, nil}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *sel) WithDistinct() sqlite.SQSelect {
	return &sel{this.source, true, this.limit, this.offset, this.where, this.to}
}

func (this *sel) WithLimitOffset(limit, offset uint) sqlite.SQSelect {
	return &sel{this.source, this.distinct, limit, offset, this.where, this.to}
}

func (this *sel) Where(v ...interface{}) sqlite.SQSelect {
	if len(v) == 0 {
		// Reset where clause
		return &sel{this.source, this.distinct, this.limit, this.offset, nil, this.to}
	} else {
		// Where clause with an expression
		return &sel{this.source, this.distinct, this.limit, this.offset, append(this.where, v...), this.to}
	}
}

func (this *sel) To(v ...sqlite.SQSource) sqlite.SQSelect {
	if len(v) == 0 {
		// Reset to clause
		return &sel{this.source, this.distinct, this.limit, this.offset, this.where, nil}
	} else {
		// To clause with an expression
		return &sel{this.source, this.distinct, this.limit, this.offset, this.where, append(this.to, v...)}
	}
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
		for i, source := range this.to {
			if i > 0 {
				token += ","
			}
			token += fmt.Sprint(source)
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
			token += fmt.Sprint(source)
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
			tokens = append(tokens, fmt.Sprint(expr))
		}
	}

	// Add offset and limit
	if this.limit == 0 && this.offset > 0 {
		tokens = append(tokens, "OFFSET", fmt.Sprint(this.offset))
	} else if this.limit > 0 && this.offset == 0 {
		tokens = append(tokens, "LIMIT", fmt.Sprint(this.limit))
	} else if this.limit > 0 && this.offset > 0 {
		tokens = append(tokens, "LIMIT", fmt.Sprint(this.offset)+","+fmt.Sprint(this.limit))
	}

	// Return the query
	return strings.Join(tokens, " ")
}
