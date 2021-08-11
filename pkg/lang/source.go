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
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// N defines a table name or column name
func N(s string) sqlite.SQSource {
	return &source{s, "", ""}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *source) WithSchema(schema string) sqlite.SQSource {
	return &source{this.name, schema, this.alias}
}

func (this *source) WithAlias(alias string) sqlite.SQSource {
	return &source{this.name, this.schema, alias}
}

func (this *source) WithType(decltype string) sqlite.SQColumn {
	return &column{*this, decltype, false, false}
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
	return strings.Join(tokens, "")
}

func (this *source) Query() string {
	return "SELECT * FROM " + this.String()
}
