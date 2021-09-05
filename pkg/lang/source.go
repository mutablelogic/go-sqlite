package lang

import (
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
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
func N(s string) sqlite.SQSource {
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

func (this *source) WithName(name string) sqlite.SQSource {
	return &source{name, this.schema, this.alias, this.desc}
}

func (this *source) WithSchema(schema string) sqlite.SQSource {
	return &source{this.name, schema, this.alias, this.desc}
}

func (this *source) WithAlias(alias string) sqlite.SQSource {
	return &source{this.name, this.schema, alias, this.desc}
}

func (this *source) WithType(decltype string) sqlite.SQColumn {
	return &column{*this, decltype, false, false, false, nil}
}

func (this *source) WithDesc() sqlite.SQSource {
	return &source{this.name, this.schema, this.alias, true}
}

///////////////////////////////////////////////////////////////////////////////
// CONVERT TO EXPR

func (this *source) Or(v interface{}) sqlite.SQExpr {
	return &e{this, v, "OR"}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *source) String() string {
	tokens := []string{}
	if this.schema != "" {
		tokens = append(tokens, sqlite.QuoteIdentifier(this.schema), ".", sqlite.QuoteIdentifier(this.name))
	} else {
		tokens = append(tokens, sqlite.QuoteIdentifier(this.name))
	}
	if this.alias != "" {
		tokens = append(tokens, " AS ", sqlite.QuoteIdentifier(this.alias))
	}
	if this.desc {
		tokens = append(tokens, " DESC")
	}
	return strings.Join(tokens, "")
}

func (this *source) Query() string {
	return this.String()
}
