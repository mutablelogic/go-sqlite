package lang

import (
	"fmt"
	"strings"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type createtable struct {
	source
	temporary    bool
	ifnotexists  bool
	withoutrowid bool
	unique       []string
	index        []string
	foreignkeys  []string
	columns      []sqlite.SQColumn
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new table with name and defined columns
func (this *source) CreateTable(columns ...sqlite.SQColumn) sqlite.SQTable {
	return &createtable{source{this.name, this.schema, "", false}, false, false, false, nil, nil, nil, columns}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *createtable) IfNotExists() sqlite.SQTable {
	return &createtable{this.source, this.temporary, true, this.ifnotexists, this.unique, this.index, this.foreignkeys, this.columns}
}

func (this *createtable) WithTemporary() sqlite.SQTable {
	return &createtable{this.source, true, this.ifnotexists, this.withoutrowid, this.unique, this.index, this.foreignkeys, this.columns}
}

func (this *createtable) WithoutRowID() sqlite.SQTable {
	return &createtable{this.source, this.temporary, this.ifnotexists, true, this.unique, this.index, this.foreignkeys, this.columns}
}

func (this *createtable) WithUnique(columns ...string) sqlite.SQTable {
	return &createtable{this.source, this.temporary, this.ifnotexists, this.withoutrowid, append(this.unique, sqlite.QuoteIdentifiers(columns...)), this.index, this.foreignkeys, this.columns}
}

func (this *createtable) WithIndex(columns ...string) sqlite.SQTable {
	return &createtable{this.source, this.temporary, this.ifnotexists, this.withoutrowid, this.unique, append(this.index, sqlite.QuoteIdentifiers(columns...)), this.foreignkeys, this.columns}
}

func (this *createtable) WithForeignKey(key sqlite.SQForeignKey, columns ...string) sqlite.SQTable {
	return &createtable{this.source, this.temporary, this.ifnotexists, this.withoutrowid, this.unique, this.index,
		append(this.foreignkeys, key.(*foreignkey).Query(columns...)), this.columns}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *createtable) String() string {
	return this.Query()
}

func (this *createtable) Query() string {
	tokens := []string{"CREATE"}
	columns := make([]string, len(this.columns), len(this.columns)+len(this.unique)+len(this.index)+1)

	// Set the columns
	primary := []string{}
	j := 0
	for i, col := range this.columns {
		if col, ok := col.(*column); ok {
			columns[i] = col.String()
			if col.primary {
				j = i
				primary = append(primary, col.name)
			}
		}
	}

	// Add primary key
	if len(primary) == 1 {
		columns[j] = fmt.Sprint(this.columns[j], " ", this.columns[j].Primary())
	} else if len(primary) > 1 {
		columns = append(columns, "PRIMARY KEY ("+sqlite.QuoteIdentifiers(primary...)+")")
	}

	// Add indexes
	if len(this.columns) > 0 {
		for _, key := range this.unique {
			columns = append(columns, "UNIQUE ("+key+")")
		}
		for _, key := range this.index {
			columns = append(columns, "INDEX ("+key+")")
		}
		columns = append(columns, this.foreignkeys...)
	}

	// Add keywords into the query
	if this.temporary {
		tokens = append(tokens, "TEMPORARY")
	}
	if this.ifnotexists {
		tokens = append(tokens, "TABLE IF NOT EXISTS")
	} else {
		tokens = append(tokens, "TABLE")
	}

	// Add table name
	tokens = append(tokens, this.source.String())

	// Add columns
	tokens = append(tokens, "("+strings.Join(columns, ",")+")")

	// Final flags
	if this.withoutrowid {
		tokens = append(tokens, "WITHOUT ROWID")
	}

	// Return the query
	return strings.Join(tokens, " ")
}
