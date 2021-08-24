package lang

import (
	"fmt"
	"strings"

	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type altertable struct {
	source
	token string
	col   sqlite.SQColumn
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new table with name and defined columns
func (this *source) AlterTable() sqlite.SQAlter {
	return &altertable{source{this.name, this.schema, "", false}, "", nil}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func (this *altertable) AddColumn(col sqlite.SQColumn) sqlite.SQStatement {
	if col == nil {
		return nil
	}
	this.col = col
	this.token = "ADD"
	return this
}

func (this *altertable) DropColumn(col sqlite.SQColumn) sqlite.SQStatement {
	if col == nil {
		return nil
	}
	this.col = col
	this.token = "DROP"
	return this
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *altertable) String() string {
	return this.Query()
}

func (this *altertable) Query() string {
	tokens := []string{"ALTER TABLE", this.source.String()}

	switch this.token {
	case "ADD":
		tokens = append(tokens, "ADD COLUMN", fmt.Sprint(this.col))
	case "DROP":
		tokens = append(tokens, "DROP COLUMN", this.col.Name())
	}

	// Return the query
	return strings.Join(tokens, " ")
}
