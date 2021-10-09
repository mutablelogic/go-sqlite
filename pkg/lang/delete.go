package lang

import (
	"fmt"
	"strings"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type delete struct {
	source SQSource
	where  []interface{}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Update values in a table with a name and defined column names
func (this *source) Delete(expr ...interface{}) SQStatement {
	if len(expr) == 0 {
		return nil
	} else {
		return &delete{&source{this.name, this.schema, "", false}, expr}
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *delete) String() string {
	return this.Query()
}

func (this *delete) Query() string {
	tokens := []string{"DELETE FROM", this.source.String()}

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

	// Return the query
	return strings.Join(tokens, " ")
}
