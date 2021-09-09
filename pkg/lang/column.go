package lang

import (
	"fmt"
	"strings"

	// Import namespaces
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/quote"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type column struct {
	source
	decltype      string
	notnull       bool
	primary       bool
	autoincrement bool
	def           SQExpr
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultColumnDecltype = "TEXT"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// C defines a column name
func C(name string) SQColumn {
	return &column{source{name, "", "", false}, defaultColumnDecltype, false, false, false, nil}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *column) Name() string {
	return this.source.Name()
}

func (this *column) Type() string {
	return this.decltype
}

func (this *column) Nullable() bool {
	return this.notnull == false
}

func (this *column) Primary() string {
	if this.autoincrement {
		return "PRIMARY KEY AUTOINCREMENT"
	} else if this.primary {
		return "PRIMARY KEY"
	} else {
		return ""
	}
}

func (this *column) WithType(v string) SQColumn {
	return &column{this.source, v, this.notnull, this.primary, this.autoincrement, this.def}
}

func (this *column) WithAlias(v string) SQSource {
	return &source{this.name, "", v, false}
}

func (this *column) NotNull() SQColumn {
	return &column{this.source, this.decltype, true, this.primary, this.autoincrement, this.def}
}

func (this *column) WithPrimary() SQColumn {
	return &column{this.source, this.decltype, true, true, this.autoincrement, this.def}
}

func (this *column) WithAutoIncrement() SQColumn {
	return &column{this.source, this.decltype, true, true, true, this.def}
}

func (this *column) WithDefault(v interface{}) SQColumn {
	return &column{this.source, this.decltype, true, true, this.autoincrement, V(v)}
}

func (this *column) WithDefaultNow() SQColumn {
	return &column{this.source, this.decltype, true, true, this.autoincrement, V("CURRENT_TIMESTAMP")}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *column) Query() string {
	return this.String()
}

func (this *column) String() string {
	tokens := []string{QuoteIdentifier(this.Name())}
	if this.decltype != "" {
		tokens = append(tokens, QuoteDeclType(this.decltype))
	} else {
		tokens = append(tokens, defaultColumnDecltype)
	}
	if this.notnull {
		tokens = append(tokens, "NOT NULL")
	}
	if this.def != nil {
		tokens = append(tokens, "DEFAULT", fmt.Sprint(this.def))
	}
	return strings.Join(tokens, " ")
}
