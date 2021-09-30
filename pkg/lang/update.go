package lang

import (
	"fmt"
	"strings"

	// Import namespaces
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/quote"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type update struct {
	source   SQSource
	conflict string
	where    []interface{}
	columns  []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Update values in a table with a name and defined column names
func (this *source) Update(columns ...string) SQUpdate {
	return &update{&source{this.name, this.schema, "", false}, "", nil, columns}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *update) WithAbort() SQUpdate {
	return &update{this.source, "OR ABORT", this.where, this.columns}
}

func (this *update) WithFail() SQUpdate {
	return &update{this.source, "OR FAIL", this.where, this.columns}
}

func (this *update) WithIgnore() SQUpdate {
	return &update{this.source, "OR IGNORE", this.where, this.columns}
}

func (this *update) WithReplace() SQUpdate {
	return &update{this.source, "OR REPLACE", this.where, this.columns}
}

func (this *update) WithRollback() SQUpdate {
	return &update{this.source, "OR ROLLBACK", this.where, this.columns}
}

func (this *update) Where(v ...interface{}) SQUpdate {
	if len(v) == 0 {
		// Reset where clause
		return &update{this.source, this.conflict, nil, this.columns}
	}
	// Where clause with an expression
	return &update{this.source, this.conflict, append(this.where, v...), this.columns}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *update) String() string {
	return this.Query()
}

func (this *update) Query() string {
	tokens := []string{"UPDATE"}
	if this.conflict != "" {
		tokens = append(tokens, this.conflict)
	}

	// Add source
	tokens = append(tokens, fmt.Sprint(this.source))

	// Add set clause
	if len(this.columns) > 0 {
		cols := make([]string, 0, len(this.columns))
		for _, col := range this.columns {
			cols = append(cols, fmt.Sprint(QuoteIdentifier(col), "=?"))
		}
		tokens = append(tokens, "SET", strings.Join(cols, ", "))
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

	// Return the query
	return strings.Join(tokens, " ")
}
