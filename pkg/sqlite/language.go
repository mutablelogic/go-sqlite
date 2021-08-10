package sqlite

import (
	"fmt"
	"strings"
	"time"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type language struct{}

type query struct {
	v string
}

type expr struct {
	v interface{}
}

type name struct {
	name   string
	schema string
	alias  string
}

type column struct {
	name
	decltype string
	nullable bool
	primary  bool
}

type insert struct {
	name
	class         string
	defaultvalues bool
	columns       []string
}

type createtable struct {
	name
	temporary    bool
	ifnotexists  bool
	withoutrowid bool
	unique       []string
	index        []string
	columns      []sqlite.SQColumn
}

type createview struct {
	name
	temporary   bool
	ifnotexists bool
	sel         sqlite.SQSelect
	columns     []string
}

type drop struct {
	name
	class    string
	ifexists bool
}

type comparison struct {
	l     sqlite.SQExpr
	r     []sqlite.SQExpr
	class string
	not   bool
}

type tablename struct {
	name   string
	schema string
	alias  string
}

type createindex struct {
	tablename

	name        string
	unique      bool
	ifnotexists bool
	columns     []string
}

type sel struct {
	source        []sqlite.SQName
	distinct      bool
	offset, limit uint
	where         []sqlite.SQExpr
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	p = &name{"?", "", ""}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Q Creates a query or value
func (language) Q(v interface{}) sqlite.SQStatement {
	switch v := v.(type) {
	case string:
		if v == "" {
			return &expr{v}
		} else {
			return &query{v}
		}
	default:
		return &expr{v}
	}
}

// V Creates a value expression
func (language) V(v interface{}) sqlite.SQExpr {
	return &expr{v}
}

// N Creates a name
func (language) N(s string) sqlite.SQName {
	return &name{s, "", ""}
}

// P Creates a bound parameter
func (language) P() sqlite.SQExpr {
	return p
}

// S Creates a select statement
func (language) S(sources ...sqlite.SQName) sqlite.SQSelect {
	return &sel{sources, false, 0, 0, nil}
}

///////////////////////////////////////////////////////////////////////////////
// QUERY

func (this *query) Query() string {
	return this.v
}

///////////////////////////////////////////////////////////////////////////////
// NAME

func (this *name) WithSchema(schema string) sqlite.SQName {
	this.schema = strings.TrimSpace(schema)
	return this
}

func (this *name) WithAlias(alias string) sqlite.SQName {
	this.alias = strings.TrimSpace(alias)
	return this
}

func (this *name) WithType(decltype string) sqlite.SQColumn {
	return &column{name{this.name, this.schema, ""}, strings.TrimSpace(decltype), true, false}
}

func (this *name) Is(r sqlite.SQExpr, rr ...sqlite.SQExpr) sqlite.SQComparison {
	return &comparison{this., nil, ""}
}

func (this *name) String() string {
	// Special case for bind parameter
	if this == p {
		return "?"
	}

	tokens := []string{}
	if this.schema != "" {
		tokens = append(tokens, QuoteIdentifier(this.schema), ".", QuoteIdentifier(this.name))
	} else {
		tokens = append(tokens, QuoteIdentifier(this.name))
	}
	if this.alias != "" {
		tokens = append(tokens, " AS ", QuoteIdentifier(this.alias))
	}
	return strings.Join(tokens, "")
}

func (this *name) Query() string {
	if this == p {
		return "SELECT ?"
	} else {
		return "SELECT * FROM " + this.String()
	}
}

// Insert values into a table with a name and defined column names
func (this *name) Insert(columns ...string) sqlite.SQInsert {
	return &insert{name{this.name, this.schema, ""}, "INSERT", false, columns}
}

// Replace values into a table with a name and defined column names
func (this *name) Replace(columns ...string) sqlite.SQInsert {
	return &insert{name{this.name, this.schema, ""}, "REPLACE", false, columns}
}

// Create a new table with name and defined columns
func (this *name) CreateTable(columns ...sqlite.SQColumn) sqlite.SQTable {
	return &createtable{name{this.name, this.schema, ""}, false, false, false, nil, nil, columns}
}

// Create a new view with name and defined columns
func (this *name) CreateView(sel sqlite.SQSelect, columns ...string) sqlite.SQIndexView {
	return &createview{name{this.name, this.schema, ""}, false, false, sel, columns}
}

// Drop a table
func (this *name) DropTable() sqlite.SQDrop {
	return &drop{name{this.name, this.schema, ""}, "TABLE", false}
}

// Drop a index
func (this *name) DropIndex() sqlite.SQDrop {
	return &drop{name{this.name, this.schema, ""}, "INDEX", false}
}

// Drop a trigger
func (this *name) DropTrigger() sqlite.SQDrop {
	return &drop{name{this.name, this.schema, ""}, "TRIGGER", false}
}

// Drop a view
func (this *name) DropView() sqlite.SQDrop {
	return &drop{name{this.name, this.schema, ""}, "VIEW", false}
}

///////////////////////////////////////////////////////////////////////////////
// EXPRESSION

func (this *expr) String() string {
	if this.v == nil {
		return "NULL"
	}
	switch e := this.v.(type) {
	case string:
		return Quote(e)
	case uint, int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(this.v)
	case bool:
		if e {
			return "TRUE"
		} else {
			return "FALSE"
		}
	case time.Time:
		if e.IsZero() {
			return "NULL"
		} else {
			return Quote(e.Format(time.RFC3339Nano))
		}
	default:
		return Quote(fmt.Sprint(this.v))
	}
}

func (this *expr) Query() string {
	return "SELECT " + this.String()
}

///////////////////////////////////////////////////////////////////////////////
// COLUMN

func (this *column) NotNull() sqlite.SQColumn {
	this.nullable = false
	return this
}

func (this *column) Primary() sqlite.SQColumn {
	this.primary = true
	this.nullable = false
	return this
}

func (this *column) String() string {
	if this.nullable {
		return fmt.Sprintf("%v %v", QuoteIdentifier(this.name.String()), QuoteDeclType(this.decltype))
	} else {
		return fmt.Sprintf("%v %v NOT NULL", QuoteIdentifier(this.name.String()), QuoteDeclType(this.decltype))
	}
}

////////////////////////////////////////////////////////////////////////////////
// INSERT OR REPLACE

func (this *insert) DefaultValues() sqlite.SQInsert {
	this.defaultvalues = true
	return this
}

func (this *insert) Query() string {
	tokens := []string{this.class, "INTO"}

	// Add table name
	tokens = append(tokens, this.name.String())

	// Add column names
	if len(this.columns) > 0 {
		tokens = append(tokens, "("+QuoteIdentifiers(this.columns...)+")")
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

func (this *insert) argsN(n int) string {
	if n < 1 {
		return ""
	} else {
		return "(" + strings.Repeat("?,", n-1) + "?)"
	}
}

////////////////////////////////////////////////////////////////////////////////
// TABLE

func (this *createtable) IfNotExists() sqlite.SQTable {
	this.ifnotexists = true
	return this
}

func (this *createtable) WithTemporary() sqlite.SQTable {
	this.temporary = true
	return this
}

func (this *createtable) WithoutRowID() sqlite.SQTable {
	this.withoutrowid = true
	return this
}

func (this *createtable) WithUnique(columns ...string) sqlite.SQTable {
	if len(columns) > 0 {
		this.unique = append(this.unique, QuoteIdentifiers(columns...))
	}
	return this
}

func (this *createtable) WithIndex(columns ...string) sqlite.SQTable {
	if len(columns) > 0 {
		this.index = append(this.index, QuoteIdentifiers(columns...))
	}
	return this
}

func (this *createtable) String() string {
	return this.Query()
}

func (this *createtable) Query() string {
	tokens := []string{"CREATE"}
	columns := make([]string, len(this.columns), len(this.columns)+len(this.unique)+len(this.index)+1)

	// Set the columns
	primary := []string{}
	for i, col := range this.columns {
		if col, ok := col.(*column); ok {
			columns[i] = col.String()
			if col.primary {
				primary = append(primary, col.name.name)
			}
		}
	}

	// Add primary key
	if len(primary) > 0 {
		columns = append(columns, "PRIMARY KEY ("+QuoteIdentifiers(primary...)+")")
	}

	// Add indexes
	for _, key := range this.unique {
		columns = append(columns, "UNIQUE ("+key+")")
	}
	for _, key := range this.index {
		columns = append(columns, "INDEX ("+key+")")
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
	tokens = append(tokens, this.name.String())

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
// DROP

func (this *drop) IfExists() sqlite.SQDrop {
	this.ifexists = true
	return this
}

func (this *drop) Query() string {
	tokens := []string{"DROP", this.class}
	if this.ifexists {
		tokens = append(tokens, "IF EXISTS")
	}
	tokens = append(tokens, this.name.String())

	// Return the query
	return strings.Join(tokens, " ")
}

func (this *drop) String() string {
	return this.Query()
}

////////////////////////////////////////////////////////////////////////////////
// SELECT

func (this *sel) WithDistinct() sqlite.SQSelect {
	this.distinct = true
	return this
}

func (this *sel) WithLimitOffset(limit, offset uint) sqlite.SQSelect {
	this.offset, this.limit = offset, limit
	return this
}

func (this *sel) Where(v ...interface{}) sqlite.SQSelect {
	if len(v) == 0 {
		// Reset where clause
		this.where = nil
	} else {
		// Append to where clause
		for _, v := range v {
			switch v := v.(type) {
			case sqlite.SQExpr:
				this.where = append(this.where, v)
			default:
				this.where = append(this.where, &expr{v})
			}
		}
	}
	return this
}

func (this *sel) String() string {
	return this.Query()
}

func (this *sel) Query() string {
	tokens := []string{"SELECT"}

	// Add distinct keyword
	if this.distinct {
		tokens = append(tokens, "DISTINCT")
	}

	// TODO: Add column expressions
	tokens = append(tokens, "*")

	// Add sources using a cross join
	if len(this.source) > 0 {
		token := "FROM "
		for i, source := range this.source {
			if i > 0 {
				token += ","
			}
			token += fmt.Sprint(source)
		}
		tokens = append(tokens, token)
	}

	// Where clause
	if len(this.where) > 0 {
		tokens = append(tokens, "WHERE")
		for i, expr := range this.where {
			if i > 0 {
				tokens = append(tokens, "AND")
			}
			tokens = append(tokens, fmt.Sprint(expr))
		}
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

///////////////////////////////////////////////////////////////////////////////
// CREATE VIEW

func (this *createview) IfNotExists() sqlite.SQIndexView {
	this.ifnotexists = true
	return this
}

func (this *createview) WithTemporary() sqlite.SQIndexView {
	this.temporary = true
	return this
}

func (this *createview) WithUnique() sqlite.SQIndexView {
	return this
}

func (this *createview) String() string {
	return this.Query()
}

func (this *createview) Query() string {
	tokens := []string{"CREATE"}
	if this.temporary {
		tokens = append(tokens, "TEMPORARY")
	}
	if this.ifnotexists {
		tokens = append(tokens, "VIEW IF NOT EXISTS")
	} else {
		tokens = append(tokens, "VIEW")
	}

	tokens = append(tokens, this.name.String())
	if len(this.columns) > 0 {
		tokens = append(tokens, "("+QuoteIdentifiers(this.columns...)+")")
	}
	tokens = append(tokens, "AS", this.sel.Query())

	// Return the query
	return strings.Join(tokens, " ")
}

/*

// Create a new index with a source table name and defined column names
func (language) CreateIndex(name string, columns ...string) sqlite.SQIndexView {
	return &createindex{tablename{"", "", ""}, name, false, false, columns}
}

// Create a new column name and declared type
func (language) Column(name, decltype string) sqlite.SQColumn {
	return &column{
		name, decltype, true, false,
	}
}

// Is creates a comparison expression
func (language) Is(l, r sqlite.SQExpr, rr ...sqlite.SQExpr) sqlite.SQComparison {
	if l == nil {
		l = &expr{}
	}
	if r == nil {
		r = &expr{}
	}
	return &comparison{l, append([]sqlite.SQExpr{r}, rr...), "=", false}
}

///////////////////////////////////////////////////////////////////////////////
// CREATE INDEX

func (this *createindex) IfNotExists() sqlite.SQIndexView {
	this.ifnotexists = true
	return this
}

func (this *createindex) WithSchema(schema string) sqlite.SQIndexView {
	this.tablename.WithSchema(schema)
	return this
}

func (this *createindex) WithUnique() sqlite.SQIndexView {
	this.unique = true
	return this
}

func (this *createindex) WithName(name string) sqlite.SQIndexView {
	this.tablename.name = name
	return this
}

func (this *createindex) WithTemporary() sqlite.SQIndexView {
	// Ignore
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
	if this.tablename.name == "" {
		// Set table index name
		this.tablename.name = strings.Join(append([]string{this.name}, this.columns...), "_")
	}
	tokens = append(tokens, this.tablename.Query(), "ON", QuoteIdentifier(this.name), "("+QuoteIdentifiers(this.columns...)+")")

	// Return the query
	return strings.Join(tokens, " ")
}

////////////////////////////////////////////////////////////////////////////////
// TABLE NAME

func (this *tablename) WithSchema(schema string) sqlite.SQSource {
	this.schema = strings.TrimSpace(schema)
	return this
}

func (this *tablename) WithAlias(alias string) sqlite.SQSource {
	this.alias = strings.TrimSpace(alias)
	return this
}

func (this *tablename) Query() string {
	if this.schema != "" {
		return QuoteIdentifier(this.schema) + "." + QuoteIdentifier(this.name)
	} else {
		return QuoteIdentifier(this.name)
	}
}

func (this *tablename) String() string {
	if this.alias != "" {
		return this.Query() + " AS " + QuoteIdentifier(this.alias)
	} else {
		return this.Query()
	}
}

///////////////////////////////////////////////////////////////////////////////
// DROP

func (this *drop) WithSchema(schema string) sqlite.SQDrop {
	this.tablename.WithSchema(schema)
	return this
}

///////////////////////////////////////////////////////////////////////////////
// COLUMN

///////////////////////////////////////////////////////////////////////////////
// VARIABLE

func (this v) Query() string {
	return fmt.Sprint("SELECT ", this)
}

func (this v) String() string {
	switch this {
	case "":
		return "NULL"
	case "?":
		return "?"
	default:
		return QuoteIdentifier(string(this))
	}
}

///////////////////////////////////////////////////////////////////////////////
// EXPRESSION

func (this *expr) Query() string {
	if this.v == nil {
		return "SELECT NULL"
	}
	switch e := this.v.(type) {
	case string:
		e = strings.TrimSpace(e)
		if e == "" {
			return "SELECT " + Quote(e)
		} else {
			return e
		}
	case uint, int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64, bool:
		return "SELECT " + fmt.Sprint(this.v)
	default:
		return "SELECT " + Quote(fmt.Sprint(this.v))
	}
}

func (this *expr) String() string {
	if this.v == nil {
		return "NULL"
	}
	switch e := this.v.(type) {
	case string:
		return Quote(e)
	case uint, int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(this.v)
	case bool:
		if e {
			return "TRUE"
		} else {
			return "FALSE"
		}
	case time.Time:
		if e.IsZero() {
			return "NULL"
		} else {
			return Quote(e.Format(time.RFC3339Nano))
		}
	default:
		return Quote(fmt.Sprint(this.v))
	}
}

///////////////////////////////////////////////////////////////////////////////
// COMPARISON

func (this *comparison) Not() sqlite.SQComparison {
	this.not = !this.not
	return this
}

func (this *comparison) Query() string {
	// l IS NULL
	if (this.class == "=") && len(this.r) == 1 && fmt.Sprint(this.r) == "NULL" {
		if this.not {
			return fmt.Sprint(this.l, " NOTNULL")
		} else {
			return fmt.Sprint(this.l, " ISNULL")
		}
	}
	// l <op> r
	if len(this.r) == 1 {
		switch this.class {
		case "=":
			if this.not {
				return fmt.Sprint(this.l, " <> ", this.r)
			} else {
				return fmt.Sprint(this.l, " = ", this.r)
			}
		default:
			if this.not {
				return fmt.Sprint("NOT(", this.l, " ", this.class, " ", this.r, ")")
			} else {
				return fmt.Sprint(this.l, " ", this.class, " ", this.r)
			}
		}
	}
	// l <op> [ r ]
	str := "("
	for i, r := range this.r {
		if i > 1 {
			str += " OR "
		}
		str += fmt.Sprint(this.l, this.class, r)
	}
	return str + ")"
}
*/
