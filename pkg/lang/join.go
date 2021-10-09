package lang

import (
	// Namespace Imports

	"strings"

	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type join struct {
	l, r  SQSource
	class string
	expr  []SQExpr
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// J defines a join between two sources
func J(l, r SQSource) SQJoin {
	return &join{l, r, "CROSS JOIN", nil}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (j *join) String() string {
	tokens := []string{j.l.String(), j.class, j.r.String()}

	// Add the join expressions
	if len(j.expr) > 0 {
		tokens = append(tokens, "ON", sliceJoin(j.expr, " AND ", nil))
	}

	// Return the join
	return strings.Join(tokens, " ")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (j *join) Join(expr ...SQExpr) SQJoin {
	return &join{j.l, j.r, "JOIN", expr}
}

func (j *join) LeftJoin(expr ...SQExpr) SQJoin {
	return &join{j.l, j.r, "LEFT JOIN", expr}
}

func (j *join) LeftInnerJoin(expr ...SQExpr) SQJoin {
	return &join{j.l, j.r, "LEFT INNER JOIN", expr}
}
