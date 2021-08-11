package lang

import (
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type column struct {
	source
	decltype string
	notnull  bool
	primary  bool
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// C defines a column name
func C(name string) sqlite.SQColumn {
	return &column{source{name, "", ""}, "TEXT", false, false}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *column) WithType(v string) sqlite.SQColumn {
	return &column{this.source, v, this.notnull, this.primary}
}

func (this *column) WithAlias(v string) sqlite.SQSource {
	return &source{this.name, "", v}
}

func (this *column) NotNull() sqlite.SQColumn {
	return &column{this.source, this.decltype, true, this.primary}
}

func (this *column) Primary() sqlite.SQColumn {
	return &column{this.source, this.decltype, true, true}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *column) String() string {
	tokens := []string{sqlite.QuoteIdentifier(this.source.String())}
	if this.decltype != "" {
		tokens = append(tokens, sqlite.QuoteDeclType(this.decltype))
	} else {
		tokens = append(tokens, "TEXT")
	}
	if this.notnull {
		tokens = append(tokens, "NOT NULL")
	}
	return strings.Join(tokens, " ")
}
