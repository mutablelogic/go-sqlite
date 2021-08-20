package lang

import (
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type createindex struct {
	source
	name        string
	unique      bool
	ifnotexists bool
	columns     []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new index with name and defined columns
func (this *source) CreateIndex(name string, columns ...string) sqlite.SQIndexView {
	return &createindex{source{this.name, this.schema, ""}, name, false, false, columns}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *createindex) IfNotExists() sqlite.SQIndexView {
	return &createindex{this.source, this.name, this.unique, true, this.columns}
}

func (this *createindex) WithUnique() sqlite.SQIndexView {
	return &createindex{this.source, this.name, true, this.ifnotexists, this.columns}
}

func (this *createindex) WithTemporary() sqlite.SQIndexView {
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *createindex) String() string {
	return this.Query()
}

func (this *createindex) Query() string {
	tokens := []string{"CREATE"}
	if this.unique {
		tokens = append(tokens, "UNIQUE INDEX")
	} else {
		tokens = append(tokens, "INDEX")
	}
	if this.ifnotexists {
		tokens = append(tokens, "IF NOT EXISTS")
	}
	tokens = append(tokens, this.source.String(), "ON", sqlite.QuoteIdentifier(this.name), "("+sqlite.QuoteIdentifiers(this.columns...)+")")

	// Return the query
	return strings.Join(tokens, " ")
}
