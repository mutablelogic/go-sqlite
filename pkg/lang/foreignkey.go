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

type foreignkey struct {
	*source
	columns    []string
	constraint string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a foreign key
func (this *source) ForeignKey(columns ...string) SQForeignKey {
	return &foreignkey{&source{this.name, "", "", false}, columns, ""}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *foreignkey) OnDeleteCascade() SQForeignKey {
	return &foreignkey{this.source, this.columns, "ON DELETE CASCADE"}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *foreignkey) String() string {
	return this.Query("<col>", "<col>")
}

func (this *foreignkey) Query(columns ...string) string {
	tokens := []string{"FOREIGN KEY (" + QuoteIdentifiers(columns...) + ")", "REFERENCES", fmt.Sprint(this.source)}

	// Add columns
	if len(this.columns) > 0 {
		tokens = append(tokens, "("+QuoteIdentifiers(this.columns...)+")")
	}

	// Add constraint clause
	if this.constraint != "" {
		tokens = append(tokens, this.constraint)
	}

	// Return the query
	return strings.Join(tokens, " ")
}
