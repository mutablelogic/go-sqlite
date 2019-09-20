/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	"fmt"
	"strconv"
	"strings"

	sq "github.com/djthorpe/sqlite"
	driver "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type statement struct {
	prepared

	query string
}

type tablename struct {
	name   string
	schema string
}

type createtable struct {
	prepared
	tablename

	temporary    bool
	ifnotexists  bool
	withoutrowid bool
	primarykey   string
	unique       []string
	columns      []sq.Column
}

type droptable struct {
	prepared
	tablename

	ifexists bool
}

type insertreplace struct {
	prepared
	tablename

	defaultvalues bool
	columns       []string
}

type tableinfo struct {
	prepared
	tablename
}

type query struct {
	prepared

	source        sq.Source
	distinct      bool
	offset, limit uint
}

type source struct {
	tablename

	alias string
}

////////////////////////////////////////////////////////////////////////////////
// NEW STATEMENTS

func (this *sqlite) NewColumn(name, decltype string, nullable bool) sq.Column {
	this.log.Debug2("<sqlite.NewColumn>{ name=%v decltype=%v nullable=%v }", strconv.Quote(name), strconv.Quote(decltype), nullable)

	if this.conn == nil {
		return nil
	} else if name = strings.TrimSpace(name); name == "" {
		return nil
	} else if decltype = strings.TrimSpace(decltype); decltype == "" {
		return nil
	} else {
		return &column{
			name, decltype, nullable, -1,
		}
	}
}

func (this *sqlite) NewStatement(query string) sq.Statement {
	this.log.Debug2("<sqlite.NewStatement>{ %v }", strconv.Quote(query))

	if this.conn == nil {
		return nil
	} else if query == "" {
		return nil
	} else {
		return &statement{prepared{nil}, query}
	}
}

func (this *sqlite) NewCreateTable(name string, columns ...sq.Column) sq.CreateTable {
	this.log.Debug2("<sqlite.NewCreateTable>{ name=%v columns=%v }", strconv.Quote(name), columns)

	if this.conn == nil {
		return nil
	} else if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &createtable{
			prepared{nil}, tablename{name, ""}, false, false, false, "", nil, columns,
		}
	}
}

func (this *sqlite) NewDropTable(name string) sq.DropTable {
	this.log.Debug2("<sqlite.NewDropTable>{ name=%v }", strconv.Quote(name))

	if this.conn == nil {
		return nil
	} else if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &droptable{
			prepared{nil}, tablename{name, ""}, false,
		}
	}
}

func (this *sqlite) NewInsert(name string, columns ...string) sq.InsertOrReplace {
	this.log.Debug2("<sqlite.NewInsert>{ name=%v columns=%v }", strconv.Quote(name), columns)

	if this.conn == nil {
		return nil
	} else if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &insertreplace{
			prepared{nil}, tablename{name, ""}, false, columns,
		}
	}
}

func (this *sqlite) NewTableInfo(name, schema string) sq.Statement {
	this.log.Debug2("<sqlite.NewTableInfo>{ name=%v schema=%v }", strconv.Quote(name), strconv.Quote(schema))

	if this.conn == nil {
		return nil
	} else if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &tableinfo{
			prepared{nil}, tablename{name, strings.TrimSpace(schema)},
		}
	}
}

func (this *sqlite) NewSelect(source sq.Source) sq.Select {
	this.log.Debug2("<sqlite.NewSelect>{ source=%v }", source)
	if this.conn == nil {
		return nil
	} else {
		return &query{prepared{nil}, source, false, 0, 0}
	}
}

func (this *sqlite) NewSource(name string) sq.Source {
	this.log.Debug2("<sqlite.NewSource>{ name=%v }", strconv.Quote(name))
	if this.conn == nil {
		return nil
	} else if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &source{
			tablename{name, ""}, "",
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// STATEMENT IMPLEMENTATION

func (this *tablename) Schema(schema string) {
	this.schema = strings.TrimSpace(schema)
}

func (this *tablename) Query() string {
	if this.schema != "" {
		return sq.QuoteIdentifier(this.schema) + "." + sq.QuoteIdentifier(this.name)
	} else {
		return sq.QuoteIdentifier(this.name)
	}
}

func (this *statement) Query(sq.Connection) string {
	return this.query
}

func (this *prepared) Stmt() *driver.SQLiteStmt {
	return this.SQLiteStmt
}

func (this *prepared) SetStmt(st *driver.SQLiteStmt) {
	this.SQLiteStmt = st
}

////////////////////////////////////////////////////////////////////////////////
// CREATE TABLE IMPLEMENTATION

func (this *createtable) Schema(schema string) sq.CreateTable {
	this.tablename.Schema(schema)
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

func (this *createtable) Query(sq.Connection) string {
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

	// Add table name
	tokens = append(tokens, this.tablename.Query())

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
	this.tablename.Schema(schema)
	return this
}

func (this *droptable) IfExists() sq.DropTable {
	this.ifexists = true
	return this
}

func (this *droptable) Query(sq.Connection) string {
	tokens := []string{"DROP TABLE"}

	// Add flags
	if this.ifexists {
		tokens = append(tokens, "IF EXISTS")
	}

	// Add table name
	tokens = append(tokens, this.tablename.Query())

	// Return the query
	return strings.Join(tokens, " ")
}

////////////////////////////////////////////////////////////////////////////////
// INSERT

func (this *insertreplace) Schema(schema string) sq.InsertOrReplace {
	this.tablename.Schema(schema)
	return this
}

func (this *insertreplace) DefaultValues() sq.InsertOrReplace {
	this.defaultvalues = true
	return this
}

func (this *insertreplace) Query(conn sq.Connection) string {
	tokens := []string{"INSERT INTO"}

	// Add table name
	tokens = append(tokens, this.tablename.Query())

	// Add column names
	if len(this.columns) > 0 {
		tokens = append(tokens, "("+sq.QuoteIdentifiers(this.columns...)+")")
	}

	// If default values
	if this.defaultvalues || (len(this.columns) == 0 && conn == nil) {
		tokens = append(tokens, "DEFAULT VALUES")
	} else if len(this.columns) > 0 {
		tokens = append(tokens, "VALUES", this.argsN(len(this.columns)))
	} else if columns, err := conn.ColumnsForTable(this.tablename.name, this.tablename.schema); err != nil {
		// Error returned
		return ""
	} else if len(columns) == 0 {
		// Table not found
		return ""
	} else {
		tokens = append(tokens, "VALUES", this.argsN(len(columns)))
	}

	// Return the query
	return strings.Join(tokens, " ")
}

func (this *insertreplace) argsN(n int) string {
	if n < 1 {
		return ""
	} else {
		return "(" + strings.Repeat("?,", n-1) + "?)"
	}
}

////////////////////////////////////////////////////////////////////////////////
// TABLE INFO

func (this *tableinfo) Query(sq.Connection) string {
	//PRAGMA schema.table_info(table-name);
	tokens := "PRAGMA "
	if this.tablename.schema != "" {
		tokens += sq.QuoteIdentifier(this.tablename.schema) + "."
	}
	tokens += "table_info(" + sq.QuoteIdentifier(this.tablename.name) + ")"
	return tokens
}

////////////////////////////////////////////////////////////////////////////////
// DATA SOURCE IMPLEMENTATION

func (this *source) Schema(schema string) sq.Source {
	this.tablename.Schema(schema)
	return this
}

func (this *source) Alias(alias string) sq.Source {
	this.alias = strings.TrimSpace(alias)
	return this
}

func (this *source) Query(sq.Connection) string {
	if this.alias == "" {
		return this.tablename.Query()
	} else {
		return this.tablename.Query() + " AS " + sq.QuoteIdentifier(this.alias)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SELECT IMPLEMENTATION

func (this *query) Distinct() sq.Select {
	this.distinct = true
	return this
}

func (this *query) LimitOffset(limit, offset uint) sq.Select {
	this.offset, this.limit = offset, limit
	return this
}

func (this *query) Query(conn sq.Connection) string {
	tokens := []string{"SELECT"}

	// Add distinct keyword
	if this.distinct {
		tokens = append(tokens, "DISTINCT")
	}

	// Add column expressions
	// TODO
	tokens = append(tokens, "*")

	// Add source
	if this.source != nil {
		tokens = append(tokens, "FROM", this.source.Query(conn))
	}

	// Add offset and limit
	if this.limit == 0 && this.offset > 0 {
		tokens = append(tokens, "OFFSET", fmt.Sprint(this.offset))
	} else if this.limit > 0 && this.offset == 0 {
		tokens = append(tokens, "LIMIT", fmt.Sprint(this.limit))
	} else if this.limit > 0 && this.offset > 0 {
		tokens = append(tokens, "LIMIT", fmt.Sprint(this.limit)+","+fmt.Sprint(this.offset))
	}

	// Return the query
	return strings.Join(tokens, " ")
}
