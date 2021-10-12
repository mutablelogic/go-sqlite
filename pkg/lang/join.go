package lang

import (
	"strings"

	// Namespace Imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/quote"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type join struct {
	l, r  SQSource
	class string
	expr  []SQExpr
	cols  []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// J defines a join between two sources
func J(l, r SQSource) SQJoin {
	return &join{l, r, "CROSS JOIN", nil, nil}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (j *join) String() string {
	tokens := []string{j.l.String(), j.class, j.r.String()}

	// Add the join expressions
	if len(j.expr) > 0 {
		tokens = append(tokens, "ON", sliceJoin(j.expr, " AND ", nil))
	}
	if len(j.cols) > 0 {
		tokens = append(tokens, "USING", "("+QuoteIdentifiers(j.cols...)+")")
	}

	// Return the join
	return strings.Join(tokens, " ")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (j *join) Join(expr ...SQExpr) SQJoin {
	return &join{j.l, j.r, "JOIN", expr, j.cols}
}

func (j *join) LeftJoin(expr ...SQExpr) SQJoin {
	return &join{j.l, j.r, "LEFT JOIN", expr, j.cols}
}

func (j *join) LeftInnerJoin(expr ...SQExpr) SQJoin {
	return &join{j.l, j.r, "LEFT INNER JOIN", expr, j.cols}
}

func (j *join) Using(cols ...string) SQJoin {
	return &join{j.l, j.r, j.class, nil, cols}
}
