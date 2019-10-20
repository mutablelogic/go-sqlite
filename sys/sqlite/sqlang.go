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

	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
	driver "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Language struct {
	// No configuration
}

type sqlang struct {
	log gopi.Logger
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
	unique       []string
	columns      []sq.Column
}

type createindex struct {
	prepared
	tablename

	name        string
	unique      bool
	ifnotexists bool
	columns     []string
}

type drop struct {
	prepared
	tablename

	token    string
	ifexists bool
}

type update struct {
	prepared
	tablename

	columns []string
	expr    sq.Expression
}

type delete struct {
	prepared
	tablename

	expr sq.Expression
}

type insertreplace struct {
	prepared
	tablename

	token         string
	defaultvalues bool
	columns       []string
}

type query struct {
	prepared

	source        sq.Source
	distinct      bool
	where         sq.Expression
	to            []sq.Expression
	as            []string
	offset, limit uint
}

type source struct {
	tablename

	alias string
}

type expr_equals struct {
	left  string
	right sq.Expression
	not   bool
}

type expr_value struct {
	query string
}

type expr_op struct {
	op   string
	expr []sq.Expression
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Language) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<sqlang.Open>{ }")

	this := new(sqlang)
	this.log = logger

	// Success
	return this, nil
}

func (this *sqlang) Close() error {
	this.log.Debug("<sqlang.Close>{ }")

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// NEW STATEMENTS

func (this *sqlang) NewCreateTable(name string, columns ...sq.Column) sq.CreateTable {
	this.log.Debug2("<sqlang.NewCreateTable>{ name=%v columns=%v }", strconv.Quote(name), columns)

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &createtable{
			prepared{nil}, tablename{name, ""}, false, false, false, nil, columns,
		}
	}
}

func (this *sqlang) NewCreateIndex(name, table string, columns ...string) sq.CreateIndex {
	this.log.Debug2("<sqlang.NewCreateTable>{ name=%v table=%v columns=%v }", strconv.Quote(name), strconv.Quote(table), columns)

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else if table = strings.TrimSpace(table); table == "" {
		return nil
	} else if len(columns) == 0 {
		return nil
	} else {
		return &createindex{
			prepared{nil}, tablename{name, ""}, table, false, false, columns,
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// DROP

func (this *sqlang) DropTable(name string) sq.Drop {
	this.log.Debug2("<sqlang.DropTable>{ name=%v }", strconv.Quote(name))

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &drop{
			prepared{nil}, tablename{name, ""}, "TABLE", false,
		}
	}
}

func (this *sqlang) DropIndex(name string) sq.Drop {
	this.log.Debug2("<sqlang.DropIndex>{ name=%v }", strconv.Quote(name))

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &drop{
			prepared{nil}, tablename{name, ""}, "INDEX", false,
		}
	}
}

func (this *sqlang) DropTrigger(name string) sq.Drop {
	this.log.Debug2("<sqlang.DropTrigger>{ name=%v }", strconv.Quote(name))

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &drop{
			prepared{nil}, tablename{name, ""}, "TRIGGER", false,
		}
	}
}

func (this *sqlang) DropView(name string) sq.Drop {
	this.log.Debug2("<sqlang.DropView>{ name=%v }", strconv.Quote(name))

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &drop{
			prepared{nil}, tablename{name, ""}, "VIEW", false,
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// INSERT OR REPLACE

func (this *sqlang) Insert(name string, columns ...string) sq.InsertOrReplace {
	this.log.Debug2("<sqlang.Insert>{ name=%v columns=%v }", strconv.Quote(name), columns)

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &insertreplace{
			prepared{nil}, tablename{name, ""}, "INSERT", false, columns,
		}
	}
}

func (this *sqlang) Replace(name string, columns ...string) sq.InsertOrReplace {
	this.log.Debug2("<sqlang.Replace>{ name=%v columns=%v }", strconv.Quote(name), columns)

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &insertreplace{
			prepared{nil}, tablename{name, ""}, "REPLACE", false, columns,
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// DELETE

func (this *sqlang) NewDelete(name string) sq.Delete {
	this.log.Debug2("<sqlang.NewDelete>{ name=%v }", strconv.Quote(name))

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &delete{
			prepared{nil}, tablename{name, ""}, nil,
		}
	}
}

func (this *delete) Schema(schema string) sq.Delete {
	this.tablename.Schema(schema)
	return this
}

func (this *delete) Where(expr sq.Expression) sq.Delete {
	if expr == nil {
		return nil
	} else {
		this.expr = expr
		return this
	}
}

func (this *delete) Query() string {
	tokens := []string{"DELETE FROM"}

	// Add table name
	tokens = append(tokens, this.tablename.Query())

	// Add where clause
	if this.expr != nil {
		tokens = append(tokens, "WHERE", this.expr.Query())
	}

	// Return the query
	return strings.Join(tokens, " ")
}

////////////////////////////////////////////////////////////////////////////////
// UPDATE

func (this *sqlang) NewUpdate(name string, columns ...string) sq.Update {
	this.log.Debug2("<sqlang.NewUpdate>{ name=%v columns=%v }", strconv.Quote(name), columns)

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else if len(columns) == 0 {
		return nil
	} else {
		return &update{
			prepared{nil}, tablename{name, ""}, columns, nil,
		}
	}
}

func (this *update) Schema(schema string) sq.Update {
	this.tablename.Schema(schema)
	return this
}

func (this *update) Where(expr sq.Expression) sq.Update {
	if expr == nil {
		return nil
	} else {
		this.expr = expr
		return this
	}
}

func (this *update) Query() string {
	tokens := []string{"UPDATE"}

	// Add table name
	tokens = append(tokens, this.tablename.Query())

	// Columns
	tokens = append(tokens, "SET")
	for i, column := range this.columns {
		if i == len(this.columns)-1 {
			tokens = append(tokens, sq.QuoteIdentifier(column)+"=?")
		} else {
			tokens = append(tokens, sq.QuoteIdentifier(column)+"=?,")
		}
	}

	// Add where clause
	if this.expr != nil {
		tokens = append(tokens, "WHERE", this.expr.Query())
	}

	// Return the query
	return strings.Join(tokens, " ")
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

func (this *statement) Query() string {
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
	primary := []string{}
	for i, column := range this.columns {
		columns[i] = column.Query()
		if column.PrimaryKey() {
			primary = append(primary, column.Name())
		}
	}

	// Add primary key
	if len(primary) > 0 {
		columns = append(columns, "PRIMARY KEY ("+sq.QuoteIdentifiers(primary...)+")")
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
// CREATE INDEX IMPLEMENTATION

func (this *createindex) Schema(schema string) sq.CreateIndex {
	this.tablename.Schema(schema)
	return this
}

func (this *createindex) IfNotExists() sq.CreateIndex {
	this.ifnotexists = true
	return this
}

func (this *createindex) Unique() sq.CreateIndex {
	this.unique = true
	return this
}

func (this *createindex) Query() string {
	tokens := []string{"CREATE"}

	if this.unique {
		tokens = append(tokens, "UNIQUE INDEX")
	} else {
		tokens = append(tokens, "INDEX")
	}

	if this.ifnotexists {
		tokens = append(tokens, "IF NOT EXISTS")
	}

	// Add index name
	tokens = append(tokens, this.tablename.Query())

	// Add table name
	tokens = append(tokens, "ON", sq.QuoteIdentifier(this.name))

	// Add columns
	tokens = append(tokens, "("+sq.QuoteIdentifiers(this.columns...)+")")

	// Return the query
	return strings.Join(tokens, " ")
}

////////////////////////////////////////////////////////////////////////////////
// DROP

func (this *drop) Schema(schema string) sq.Drop {
	this.tablename.Schema(schema)
	return this
}

func (this *drop) IfExists() sq.Drop {
	this.ifexists = true
	return this
}

func (this *drop) Query() string {
	tokens := []string{"DROP", this.token}

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

func (this *insertreplace) Query() string {
	tokens := []string{this.token, "INTO"}

	// Add table name
	tokens = append(tokens, this.tablename.Query())

	// Add column names
	if len(this.columns) > 0 {
		tokens = append(tokens, "("+sq.QuoteIdentifiers(this.columns...)+")")
	}

	// If default values
	if this.defaultvalues || (len(this.columns) == 0) {
		tokens = append(tokens, "DEFAULT VALUES")
	} else if len(this.columns) > 0 {
		tokens = append(tokens, "VALUES", this.argsN(len(this.columns)))
	} else {
		// No columns, return empty query
		return ""
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
// DATA SOURCE IMPLEMENTATION

func (this *sqlang) NewSource(name string) sq.Source {
	this.log.Debug2("<sqlang.NewSource>{ name=%v }", strconv.Quote(name))

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else {
		return &source{
			tablename{name, ""}, "",
		}
	}
}

func (this *source) Schema(schema string) sq.Source {
	this.tablename.Schema(schema)
	return this
}

func (this *source) Alias(alias string) sq.Source {
	this.alias = strings.TrimSpace(alias)
	return this
}

func (this *source) Query() string {
	if this.alias == "" {
		return this.tablename.Query()
	} else {
		return this.tablename.Query() + " AS " + sq.QuoteIdentifier(this.alias)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SELECT IMPLEMENTATION

func (this *sqlang) NewSelect(source sq.Source, expr ...sq.Expression) sq.Select {
	this.log.Debug2("<sqlang.NewSelect>{ source=%v expr=%v }", source, expr)

	return &query{prepared{nil}, source, false, nil, expr, nil, 0, 0}
}

func (this *query) Where(expr ...sq.Expression) sq.Select {
	if len(expr) == 0 {
		this.where = nil
	} else if len(expr) == 1 {
		this.where = expr[0]
	} else {
		this.where = &expr_op{" AND ", expr}
	}
	return this
}

func (this *query) Distinct() sq.Select {
	this.distinct = true
	return this
}

func (this *query) As(as ...string) sq.Select {
	if len(this.to) != len(as) {
		return nil
	}
	this.as = as
	return this
}

func (this *query) LimitOffset(limit, offset uint) sq.Select {
	this.offset, this.limit = offset, limit
	return this
}

func (this *query) Query() string {
	tokens := []string{"SELECT"}

	// Add distinct keyword
	if this.distinct {
		tokens = append(tokens, "DISTINCT")
	}

	// Add column expressions
	if len(this.to) > 0 {
		expr_ := ""
		for i, expr := range this.to {
			if i < len(this.as) {
				expr_ += fmt.Sprintf("%s AS %s", expr.Query(), sq.QuoteIdentifier(this.as[i])) + ","
			} else {
				expr_ += expr.Query() + ","
			}
		}
		tokens = append(tokens, strings.TrimSuffix(expr_, ","))
	} else {
		tokens = append(tokens, "*")
	}

	// Add source
	if this.source != nil {
		tokens = append(tokens, "FROM", this.source.Query())
	}

	// Add where
	if this.where != nil {
		// Remove brackets
		q := this.where.Query()
		if strings.HasPrefix(q, "(") && strings.HasSuffix(q, ")") {
			q = strings.TrimPrefix(q, "(")
			q = strings.TrimSuffix(q, ")")
		}
		tokens = append(tokens, "WHERE", q)
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

////////////////////////////////////////////////////////////////////////////////
// VALUE EXPRESSIONS

func (this *sqlang) Null() sq.Expression {
	return &expr_value{"NULL"}
}

func (this *sqlang) Arg() sq.Expression {
	return &expr_value{"?"}
}

func (this *sqlang) Value(v interface{}) sq.Expression {
	switch v.(type) {
	case string:
		return &expr_value{sq.DoubleQuote(v.(string))}
	case int:
		return &expr_value{fmt.Sprint(v.(int))}
	case int8:
		return &expr_value{fmt.Sprint(v.(int8))}
	case int16:
		return &expr_value{fmt.Sprint(v.(int16))}
	case int32:
		return &expr_value{fmt.Sprint(v.(int32))}
	case int64:
		return &expr_value{fmt.Sprint(v.(int64))}
	case uint:
		return &expr_value{fmt.Sprint(v.(uint))}
	case uint8:
		return &expr_value{fmt.Sprint(v.(uint8))}
	case uint16:
		return &expr_value{fmt.Sprint(v.(uint16))}
	case uint32:
		return &expr_value{fmt.Sprint(v.(uint32))}
	case uint64:
		return &expr_value{fmt.Sprint(v.(uint64))}
	case bool:
		if v.(bool) {
			return &expr_value{"TRUE"}
		} else {
			return &expr_value{"FALSE"}
		}
		return &expr_value{fmt.Sprint(v.(uint64))}
	case float32:
		return &expr_value{fmt.Sprint(v.(float32))}
	case float64:
		return &expr_value{fmt.Sprint(v.(float64))}
	default:
		// Unsupported type
		return nil
	}
}

func (this *expr_value) Query() string {
	return this.query
}

////////////////////////////////////////////////////////////////////////////////
// EQUALITY EXPRESSIONS

func (this *sqlang) Equals(name string, expr sq.Expression) sq.Expression {
	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else if expr == nil {
		return nil
	} else {
		return &expr_equals{name, expr, false}
	}
}

func (this *sqlang) NotEquals(name string, expr sq.Expression) sq.Expression {
	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else if expr == nil {
		return nil
	} else {
		return &expr_equals{name, expr, true}
	}
}

func (this *expr_equals) Query() string {
	if this.right.Query() == "NULL" {
		if this.not {
			return sq.QuoteIdentifier(this.left) + " IS NOT NULL"
		} else {
			return sq.QuoteIdentifier(this.left) + " IS NULL"
		}
	} else {
		if this.not {
			return sq.QuoteIdentifier(this.left) + "!=" + this.right.Query()
		} else {
			return sq.QuoteIdentifier(this.left) + "=" + this.right.Query()
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// AND OR EXPRESSIONS

func (this *sqlang) And(expr ...sq.Expression) sq.Expression {
	if len(expr) == 0 {
		return nil
	} else if len(expr) == 1 {
		return expr[0]
	} else {
		return &expr_op{" AND ", expr}
	}
}

func (this *sqlang) Or(expr ...sq.Expression) sq.Expression {
	if len(expr) == 0 {
		return nil
	} else if len(expr) == 1 {
		return expr[0]
	} else {
		return &expr_op{" OR ", expr}
	}
}

func (this *expr_op) Query() string {
	parts := make([]string, len(this.expr))
	for i, expr := range this.expr {
		parts[i] = expr.Query()
	}
	return "(" + strings.Join(parts, this.op) + ")"
}

////////////////////////////////////////////////////////////////////////////////
// GROUPING EXPRESSIONS

func (this *sqlang) CountAll() sq.Expression {
	return &expr_value{"COUNT(*)"}
}
