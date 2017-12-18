/*
	SQLite client
	(c) Copyright David Thorpe 2017
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

/////////////////////////////////////////////////////////////////////
// TYPES

type q_CreateTable struct {
	name          string
	schema        string
	ifnotexists   bool
	temporary     bool
	without_rowid bool
	columns       []Column
}

///////////////////////////////////////////////////////////////////////////////
// CONSTRUCT QUERIES

func CreateTable(name string, columns []Column) Statement {
	return &q_CreateTable{
		name: name,
	}
}

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

func (this *q_CreateTable) Schema(name string) Statement {
	this.schema = name
	return this
}

///////////////////////////////////////////////////////////////////////////////
// IF NOT EXISTS

func (this *q_CreateTable) IfNotExists() Statement {
	this.ifnotexists = true
	return this
}

///////////////////////////////////////////////////////////////////////////////
// TEMPORARY

func (this *q_CreateTable) Temporary() Statement {
	this.temporary = true
	return this
}

func (this *q_CreateTable) WithoutRowID() Statement {
	this.without_rowid = true
	return this
}

///////////////////////////////////////////////////////////////////////////////
// CREATE TABLE

func (this *q_CreateTable) sqlTableName() string {
	if this.name == "" {
		return ""
	}
	if this.schema != "" {
		return QuoteIdentifier(this.schema) + "." + QuoteIdentifier(this.name)
	} else {
		return QuoteIdentifier(this.name)
	}
}

func (this *q_CreateTable) SQL() string {
	sql := "CREATE"
	if this.temporary {
		sql = sql + " TEMPORARY"
	}
	sql = sql + " TABLE"
	if this.ifnotexists {
		sql = sql + " IF NOT EXISTS"
	}
	sql = sql + " " + this.sqlTableName()

	// TODO: COLUMNS

	if this.without_rowid {
		sql = sql + " WITHOUT ROWID"
	}

	return sql
}
