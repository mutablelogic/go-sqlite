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
	auto        bool
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
	return &createindex{source{this.name, this.schema, "", false}, name, false, false, columns, false}
}

// Create a virtual table with module name name and arguments
func (this *source) CreateVirtualTable(module string, args ...string) sqlite.SQIndexView {
	return &createvirtual{source{this.name, this.schema, "", false}, module, false, args}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// Return whether the index is unique
func (this *createindex) Unique() bool {
	return this.unique
}
func (this *createvirtual) Unique() bool {
	return false
}

// Return the table linked to the index, or module name
func (this *createindex) Table() string {
	return this.name
}
func (this *createvirtual) Table() string {
	return this.module
}

// Return the columns of the table in this index, or module arguments
func (this *createindex) Columns() []string {
	result := make([]string, len(this.columns))
	for i := range this.columns {
		result[i] = this.columns[i]
	}
	return result
}
func (this *createvirtual) Columns() []string {
	result := make([]string, len(this.args))
	for i := range this.args {
		result[i] = this.args[i]
	}
	return result
}

// Return whether the index is automatically generated
func (this *createindex) Auto() bool {
	return this.auto
}
func (this *createvirtual) Auto() bool {
	return false
}

func (this *createindex) IfNotExists() sqlite.SQIndexView {
	return &createindex{this.source, this.name, this.unique, true, this.columns, this.auto}
}

func (this *createindex) WithUnique() sqlite.SQIndexView {
	return &createindex{this.source, this.name, true, this.ifnotexists, this.columns, this.auto}
}

func (this *createindex) WithTemporary() sqlite.SQIndexView {
	return nil
}

func (this *createindex) WithAuto() sqlite.SQIndexView {
	return &createindex{this.source, this.name, true, this.ifnotexists, this.columns, true}
}

func (this *createvirtual) IfNotExists() sqlite.SQIndexView {
	return &createvirtual{this.source, this.module, true, this.args}
}

func (this *createvirtual) WithUnique() sqlite.SQIndexView {
	return nil
}

func (this *createvirtual) WithTemporary() sqlite.SQIndexView {
	return &createvirtual{source{this.name, "temp", "", false}, this.module, this.ifnotexists, this.args}
}

func (this *createvirtual) WithAuto() sqlite.SQIndexView {
	return nil
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
	tokens := []string{}
	if this.auto {
		tokens = append(tokens, "AUTO")
	} else {
		tokens = append(tokens, "CREATE")
	}
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
