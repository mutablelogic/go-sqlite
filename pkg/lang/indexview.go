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

type createvirtual struct {
	source
	module      string
	ifnotexists bool
	args        []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new index with name and defined columns
func (this *source) CreateIndex(name string, columns ...string) sqlite.SQIndexView {
	return &createindex{source{this.name, this.schema, ""}, name, false, false, columns}
}

// Create a virtual table with module name name and arguments
func (this *source) CreateVirtualTable(module string, args ...string) sqlite.SQIndexView {
	return &createvirtual{source{this.name, this.schema, ""}, module, false, args}
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

func (this *createvirtual) IfNotExists() sqlite.SQIndexView {
	return &createvirtual{this.source, this.module, true, this.args}
}

func (this *createvirtual) WithUnique() sqlite.SQIndexView {
	return nil
}

func (this *createvirtual) WithTemporary() sqlite.SQIndexView {
	return &createvirtual{source{this.name, "temp", ""}, this.module, this.ifnotexists, this.args}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *createindex) String() string {
	return this.Query()
}

func (this *createvirtual) String() string {
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

func (this *createvirtual) Query() string {
	tokens := []string{"CREATE VIRTUAL TABLE"}
	if this.ifnotexists {
		tokens = append(tokens, "IF NOT EXISTS")
	}
	tokens = append(tokens, this.source.String(), "USING", sqlite.QuoteIdentifier(this.module))
	if len(this.args) > 0 {
		tokens = append(tokens, "("+sqlite.QuoteIdentifiers(this.args...)+")")
	}

	// Return the query
	return strings.Join(tokens, " ")
}
