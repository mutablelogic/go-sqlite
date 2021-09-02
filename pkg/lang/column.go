package lang

import (
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type column struct {
	source
	decltype      string
	notnull       bool
	primary       bool
	autoincrement bool
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultColumnDecltype = "TEXT"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// C defines a column name
func C(name string) sqlite.SQColumn {
	return &column{source{name, "", "", false}, defaultColumnDecltype, false, false, false}
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *column) Name() string {
	return this.source.String()
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

func (this *column) WithType(v string) sqlite.SQColumn {
	return &column{this.source, v, this.notnull, this.primary, this.autoincrement}
}

func (this *column) WithAlias(v string) sqlite.SQSource {
	return &source{this.name, "", v, false}
}

func (this *column) NotNull() sqlite.SQColumn {
	return &column{this.source, this.decltype, true, this.primary, this.autoincrement}
}

func (this *column) WithPrimary() sqlite.SQColumn {
	return &column{this.source, this.decltype, true, true, this.autoincrement}
}

func (this *column) WithAutoIncrement() sqlite.SQColumn {
	return &column{this.source, this.decltype, true, true, true}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *column) String() string {
	tokens := []string{this.Name()}
	if this.decltype != "" {
		tokens = append(tokens, sqlite.QuoteDeclType(this.decltype))
	} else {
		tokens = append(tokens, defaultColumnDecltype)
	}
	if this.notnull {
		tokens = append(tokens, "NOT NULL")
	}
	return strings.Join(tokens, " ")
}
