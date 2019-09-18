/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package sqlite

import (
	"strings"

	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type createtable struct {
	name         string
	schema       string
	temporary    bool
	ifnotexists  bool
	withoutrowid bool
	primarykey   string
	unique       []string
	columns      []sq.Column
}

type droptable struct {
	name     string
	schema   string
	ifexists bool
}

type insertreplace struct {
	name          string
	schema        string
	defaultvalues bool
	columns       []string
}

////////////////////////////////////////////////////////////////////////////////
// NEW STATEMENTS

func (this *sqlite) NewColumn(name, decltype string, nullable bool) sq.Column {
	return &column{
		name, decltype, nullable, -1,
	}
}

func (this *sqlite) NewCreateTable(name string, columns ...sq.Column) sq.CreateTable {
	if name = strings.TrimSpace(name); name == "" {
		return nil
	}
	return &createtable{
		name, "", false, false, false, "", nil, columns,
	}
}

func (this *sqlite) NewDropTable(name string) sq.DropTable {
	if name = strings.TrimSpace(name); name == "" {
		return nil
	}
	return &droptable{
		name, "", false,
	}
}

func (this *sqlite) NewInsert(name string, columns ...string) sq.InsertOrReplace {
	if name = strings.TrimSpace(name); name == "" {
		return nil
	}
	return &insertreplace{
		name, "", false, columns,
	}
}

////////////////////////////////////////////////////////////////////////////////
// CREATE TABLE IMPLEMENTATION

func (this *createtable) Schema(schema string) sq.CreateTable {
	this.schema = strings.TrimSpace(schema)
	return this
}

func (this *createtable) IfNotExists() sq.CreateTable {
	this.ifnotexists = true
	return this
}

func (this *createtable) Temporary() sq.CreateTable {
	this.temporary = true
	return this
}

func (this *createtable) WithoutRowID() sq.CreateTable {
	this.withoutrowid = true
	return this
}

func (this *createtable) PrimaryKey(columns ...string) sq.CreateTable {
	this.primarykey = ""
	for i, column := range columns {
		if i > 0 {
			this.primarykey += ","
		}
		this.primarykey += sq.QuoteIdentifier(column)
	}
	return this
}

func (this *createtable) Unique(columns ...string) sq.CreateTable {
	if this.unique == nil || len(columns) == 0 {
		this.unique = make([]string, 0, 1)
	}
	if len(columns) > 0 {
		keys := ""
		for i, column := range columns {
			if i > 0 {
				keys += ","
			}
			keys += sq.QuoteIdentifier(column)
		}
		this.unique = append(this.unique, keys)
	}
	return this
}

func (this *createtable) Query() string {
	tokens := []string{"CREATE"}
	columns := make([]string, len(this.columns), len(this.columns)+len(this.unique)+1)

	// Set the columns
	for i, column := range this.columns {
		columns[i] = column.Query()
	}

	// Add primary key
	if this.primarykey != "" {
		columns = append(columns, "PRIMARY KEY ("+this.primarykey+")")
	}

	// Add unique indexes
	for _, key := range this.unique {
		columns = append(columns, "UNIQUE ("+key+")")
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

	// Add table schema and name
	if this.schema != "" {
		tokens = append(tokens, sq.QuoteIdentifier(this.schema)+"."+sq.QuoteIdentifier(this.name))
	} else {
		tokens = append(tokens, sq.QuoteIdentifier(this.name))
	}

	// Add columns
	tokens = append(tokens, "("+strings.Join(columns, ",")+")")

	// Final flags
	if this.withoutrowid {
		tokens = append(tokens, "WITHOUT ROWID")
	}

	// Return the query
	return strings.Join(tokens, " ")
}

////////////////////////////////////////////////////////////////////////////////
// DROP TABLE

func (this *droptable) Schema(schema string) sq.DropTable {
	this.schema = strings.TrimSpace(schema)
	return this
}

func (this *droptable) IfExists() sq.DropTable {
	this.ifexists = true
	return this
}

func (this *droptable) Query() string {
	tokens := []string{"DROP TABLE"}

	// Add flags
	if this.ifexists {
		tokens = append(tokens, "IF EXISTS")
	}
	// Add table schema and name
	if this.schema != "" {
		tokens = append(tokens, sq.QuoteIdentifier(this.schema)+"."+sq.QuoteIdentifier(this.name))
	} else {
		tokens = append(tokens, sq.QuoteIdentifier(this.name))
	}
	// Return the query
	return strings.Join(tokens, " ")
}

////////////////////////////////////////////////////////////////////////////////
// INSERT

func (this *insertreplace) Schema(schema string) sq.InsertOrReplace {
	this.schema = strings.TrimSpace(schema)
	return this
}

func (this *insertreplace) DefaultValues() sq.InsertOrReplace {
	this.defaultvalues = true
	return this
}

func (this *insertreplace) Query() string {
	tokens := []string{"INSERT INTO"}

	// Add table schema and name
	if this.schema != "" {
		tokens = append(tokens, sq.QuoteIdentifier(this.schema)+"."+sq.QuoteIdentifier(this.name))
	} else {
		tokens = append(tokens, sq.QuoteIdentifier(this.name))
	}

	// Add column names
	if len(this.columns) > 0 {
		// TODO
		tokens = append(tokens, "()")
	}

	// If default values
	if this.defaultvalues {
		tokens = append(tokens, "DEFAULT VALUES")
	} else {
		tokens = append(tokens, "VALUES ()")
	}

	// Return the query
	return strings.Join(tokens, " ")
}
