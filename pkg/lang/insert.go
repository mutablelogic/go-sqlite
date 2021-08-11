package lang

import (
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type insert struct {
	source
	class         string
	defaultvalues bool
	columns       []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Insert values into a table with a name and defined column names
func (this *source) Insert(columns ...string) sqlite.SQInsert {
	return &insert{source{this.name, this.schema, ""}, "INSERT", false, columns}
}

// Replace values into a table with a name and defined column names
func (this *source) Replace(columns ...string) sqlite.SQInsert {
	return &insert{source{this.name, this.schema, ""}, "REPLACE", false, columns}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *insert) DefaultValues() sqlite.SQInsert {
	return &insert{this.source, this.class, true, this.columns}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *insert) String() string {
	return this.Query()
}

func (this *insert) Query() string {
	tokens := []string{this.class, "INTO"}

	// Add table name
	tokens = append(tokens, this.source.String())

	// Add column names
	if len(this.columns) > 0 {
		tokens = append(tokens, "("+sqlite.QuoteIdentifiers(this.columns...)+")")
	}

	// If default values
	if this.defaultvalues || (len(this.columns) == 0) {
		tokens = append(tokens, "DEFAULT VALUES")
	} else if len(this.columns) > 0 {
		tokens = append(tokens, "VALUES", this.argsN(len(this.columns)))
	} else {
		// No columns, return empty query
		return ""
	}

	// Return the query
	return strings.Join(tokens, " ")
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *insert) argsN(n int) string {
	if n < 1 {
		return ""
	} else {
		return "(" + strings.Repeat("?,", n-1) + "?)"
	}
}
