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
	where         []sqlite.SQExpr
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// S defines a select statement
func S(sources ...sqlite.SQSource) sqlite.SQSelect {
	return &sel{sources, false, 0, 0, nil}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *sel) WithDistinct() sqlite.SQSelect {
	return &sel{this.source, true, this.limit, this.offset, this.where}
}

func (this *sel) WithLimitOffset(limit, offset uint) sqlite.SQSelect {
	return &sel{this.source, this.distinct, limit, offset, this.where}
}

func (this *sel) Where(v ...sqlite.SQExpr) sqlite.SQSelect {
	if len(v) == 0 {
		// Reset where clause
		return &sel{this.source, this.distinct, this.limit, this.offset, nil}
	} else if len(v) == 1 {
		// Where clause with an expression
		return &sel{this.source, this.distinct, this.limit, this.offset, []sqlite.SQExpr{v[0]}}
	} else if len(v) == 2 {
		// Where clause with A OR B
		return &sel{this.source, this.distinct, this.limit, this.offset, []sqlite.SQExpr{v[0].Or(v[1])}}
	} else {
		// TODO
		panic("XXX")
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
	if len(this.source) == 0 {
		return "SELECT NULL"
	}

	// Add distinct keyword
	if this.distinct {
		tokens = append(tokens, "DISTINCT")
	}

	// TODO: Add column expressions
	tokens = append(tokens, "*")

	// Add sources using a cross join
	token := "FROM "
	for i, source := range this.source {
		if i > 0 {
			token += ","
		}
		token += fmt.Sprint(source)
	}
	tokens = append(tokens, token)

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
		tokens = append(tokens, "LIMIT", fmt.Sprint(this.limit)+","+fmt.Sprint(this.offset))
	}

	// Return the query
	return strings.Join(tokens, " ")
}
