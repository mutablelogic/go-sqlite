package lang

import (
	"fmt"
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type delete struct {
	source sqlite.SQSource
	where  []interface{}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Update values in a table with a name and defined column names
func (this *source) Delete(expr ...interface{}) sqlite.SQStatement {
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
	tokens := []string{"DELETE FROM", fmt.Sprint(this.source)}

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
