package lang

import (
	"strings"

	// Import namespaces
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/quote"
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
	opts        []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new index with name and defined columns
func (this *source) CreateIndex(name string, columns ...string) SQIndexView {
	return &createindex{source{this.name, this.schema, "", false}, name, false, false, columns, false}
}

// Create a virtual table with module name name and arguments
func (this *source) CreateVirtualTable(module string, args ...string) SQIndexView {
	return &createvirtual{source{this.name, this.schema, "", false}, module, false, args, nil}
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

func (this *createindex) IfNotExists() SQIndexView {
	return &createindex{this.source, this.name, this.unique, true, this.columns, this.auto}
}

func (this *createindex) WithUnique() SQIndexView {
	return &createindex{this.source, this.name, true, this.ifnotexists, this.columns, this.auto}
}

func (this *createindex) WithTemporary() SQIndexView {
	return nil
}

func (this *createindex) WithAuto() SQIndexView {
	return &createindex{this.source, this.name, true, this.ifnotexists, this.columns, true}
}

func (this *createindex) Options(opts ...string) SQIndexView {
	return nil
}

func (this *createvirtual) IfNotExists() SQIndexView {
	return &createvirtual{this.source, this.module, true, this.args, this.opts}
}

func (this *createvirtual) WithUnique() SQIndexView {
	return nil
}

func (this *createvirtual) WithTemporary() SQIndexView {
	return &createvirtual{source{this.name, "temp", "", false}, this.module, this.ifnotexists, this.args, this.opts}
}

func (this *createvirtual) WithAuto() SQIndexView {
	return nil
}

func (this *createvirtual) Options(opts ...string) SQIndexView {
	return &createvirtual{source{this.name, this.schema, "", false}, this.module, this.ifnotexists, this.args, opts}
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
	tokens = append(tokens, this.source.String(), "ON", QuoteIdentifier(this.name), "("+QuoteIdentifiers(this.columns...)+")")

	// Return the query
	return strings.Join(tokens, " ")
}

func (this *createvirtual) Query() string {
	tokens := []string{"CREATE VIRTUAL TABLE"}
	if this.ifnotexists {
		tokens = append(tokens, "IF NOT EXISTS")
	}
	tokens = append(tokens, this.source.String(), "USING", QuoteIdentifier(this.module))
	argsopts := []string{}
	if len(this.args) > 0 {
		argsopts = append(argsopts, QuoteIdentifiers(this.args...))
	}
	if len(this.opts) > 0 {
		argsopts = append(argsopts, strings.Join(this.opts, ","))
	}
	if len(argsopts) > 0 {
		tokens = append(tokens, "(", strings.Join(argsopts, ",")+")")
	}

	// Return the query
	return strings.Join(tokens, " ")
}
