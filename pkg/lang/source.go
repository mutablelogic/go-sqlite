package lang

import (
	"strings"

	// Import namespaces
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/quote"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type source struct {
	name   string
	schema string
	alias  string
	desc   bool
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// N defines a table name or column name
func N(s string) SQSource {
	return &source{s, "", "", false}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *source) Name() string {
	return this.name
}

func (this *source) Schema() string {
	return this.schema
}

func (this *source) Alias() string {
	return this.alias
}

func (this *source) WithName(name string) SQSource {
	return &source{name, this.schema, this.alias, this.desc}
}

func (this *source) WithSchema(schema string) SQSource {
	return &source{this.name, schema, this.alias, this.desc}
}

func (this *source) WithAlias(alias string) SQSource {
	return &source{this.name, this.schema, alias, this.desc}
}

func (this *source) WithType(decltype string) SQColumn {
	return &column{*this, decltype, false, false, false, nil}
}

func (this *source) WithDesc() SQSource {
	return &source{this.name, this.schema, this.alias, true}
}

///////////////////////////////////////////////////////////////////////////////
// CONVERT TO EXPR

func (this *source) Or(v interface{}) SQExpr {
	return &e{this, v, "OR"}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *source) String() string {
	tokens := []string{}
	if this.schema != "" {
		tokens = append(tokens, QuoteIdentifier(this.schema), ".", QuoteIdentifier(this.name))
	} else {
		tokens = append(tokens, QuoteIdentifier(this.name))
	}
	if this.alias != "" {
		tokens = append(tokens, " AS ", QuoteIdentifier(this.alias))
	}
	if this.desc {
		tokens = append(tokens, " DESC")
	}
	return strings.Join(tokens, "")
}

func (this *source) Query() string {
	return this.String()
}
