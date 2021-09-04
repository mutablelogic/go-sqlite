package lang

import (
	"fmt"
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
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
func (this *source) ForeignKey(columns ...string) sqlite.SQForeignKey {
	return &foreignkey{&source{this.name, "", "", false}, columns, ""}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *foreignkey) OnDeleteCascade() sqlite.SQForeignKey {
	return &foreignkey{this.source, this.columns, "ON DELETE CASCADE"}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *foreignkey) String() string {
	return this.Query("<col>", "<col>")
}

func (this *foreignkey) Query(columns ...string) string {
	tokens := []string{"FOREIGN KEY (" + sqlite.QuoteIdentifiers(columns...) + ")", "REFERENCES", fmt.Sprint(this.source)}

	// Add columns
	if len(this.columns) > 0 {
		tokens = append(tokens, "("+sqlite.QuoteIdentifiers(this.columns...)+")")
	}

	// Add constraint clause
	if this.constraint != "" {
		tokens = append(tokens, this.constraint)
	}

	// Return the query
	return strings.Join(tokens, " ")
}
