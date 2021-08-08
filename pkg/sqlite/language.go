package sqlite

import (
	"fmt"
	"strings"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type language struct{}

type query struct {
	q string
}

type tablename struct {
	name   string
	schema string
	alias  string
}

type column struct {
	name     string
	decltype string
	nullable bool
	primary  bool
}

type createtable struct {
	tablename

	temporary    bool
	ifnotexists  bool
	withoutrowid bool
	unique       []string
	index        []string
	columns      []sqlite.SQColumn
}

type createindex struct {
	tablename

	name        string
	unique      bool
	ifnotexists bool
	columns     []string
}

type drop struct {
	tablename
	class    string
	ifexists bool
}

type insert struct {
	tablename
	class         string
	defaultvalues bool
	columns       []string
}

type sel struct {
	source        []sqlite.SQSource
	distinct      bool
	offset, limit uint
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Q Creates an arbitary query from a string
func (language) Q(v string) sqlite.SQStatement {
	return &query{v}
}

// Create a new table with name and defined columns
func (language) CreateTable(name string, columns ...sqlite.SQColumn) sqlite.SQTable {
	return &createtable{
		tablename{name, "", ""}, false, false, false, nil, nil, columns,
	}
}

// Create a new index with a source table name and defined column names
func (language) CreateIndex(name string, columns ...string) sqlite.SQIndex {
	return &createindex{
		tablename{"", "", ""}, name, false, false, columns,
	}
}

// Create a new column name and declared type
func (language) Column(name, decltype string) sqlite.SQColumn {
	return &column{
		name, decltype, true, false,
	}
}

// Drop a table with a name
func (language) DropTable(name string) sqlite.SQDrop {
	return &drop{tablename{name, "", ""}, "TABLE", false}
}

// Drop an index with a name
func (language) DropIndex(name string) sqlite.SQDrop {
	return &drop{tablename{name, "", ""}, "INDEX", false}
}

// Drop a trigger with a name
func (language) DropTrigger(name string) sqlite.SQDrop {
	return &drop{tablename{name, "", ""}, "TRIGGER", false}
}

// Drop a view with a name
func (language) DropView(name string) sqlite.SQDrop {
	return &drop{tablename{name, "", ""}, "VIEW", false}
}

// Insert values into a table with a name and defined column names
func (language) Insert(name string, columns ...string) sqlite.SQInsert {
	return &insert{tablename{name, "", ""}, "INSERT", false, columns}
}

// Replace values into a table with a name and defined column names
func (language) Replace(name string, columns ...string) sqlite.SQInsert {
	return &insert{tablename{name, "", ""}, "REPLACE", false, columns}
}

// Select data from source tables, expressions or select statements
func (this *language) Select(sources ...sqlite.SQSource) sqlite.SQSelect {
	return &sel{sources, false, 0, 0}
}

func (this *language) TableSource(name string) sqlite.SQSource {
	return &tablename{name, "", ""}
}

///////////////////////////////////////////////////////////////////////////////
// ARBITARY QUERY

func (this *query) Query() string {
	return this.q
}

///////////////////////////////////////////////////////////////////////////////
// SELECT

func (this *sel) WithDistinct() sqlite.SQSelect {
	this.distinct = true
	return this
}

func (this *sel) WithLimitOffset(limit, offset uint) sqlite.SQSelect {
	this.offset, this.limit = offset, limit
	return this
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
// CREATE TABLE

func (this *createtable) WithSchema(schema string) sqlite.SQTable {
	this.tablename.WithSchema(schema)
	return this
}

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

func (this *createtable) Query() string {
	tokens := []string{"CREATE"}
	columns := make([]string, len(this.columns), len(this.columns)+len(this.unique)+len(this.index)+1)

	// Set the columns
	primary := []string{}
	for i, col := range this.columns {
		if col, ok := col.(*column); ok {
			columns[i] = col.Query()
			if col.primary {
				primary = append(primary, col.name)
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

///////////////////////////////////////////////////////////////////////////////
// CREATE INDEX

func (this *createindex) IfNotExists() sqlite.SQIndex {
	this.ifnotexists = true
	return this
}

func (this *createindex) WithSchema(schema string) sqlite.SQIndex {
	this.tablename.WithSchema(schema)
	return this
}

func (this *createindex) WithUnique() sqlite.SQIndex {
	this.unique = true
	return this
}

func (this *createindex) WithName(name string) sqlite.SQIndex {
	this.tablename.name = name
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

func (this *tablename) WithSchema(schema string) {
	this.schema = strings.TrimSpace(schema)
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

func (this *drop) IfExists() sqlite.SQDrop {
	this.ifexists = true
	return this
}

func (this *drop) Query() string {
	tokens := []string{"DROP", this.class}
	if this.ifexists {
		tokens = append(tokens, "IF EXISTS")
	}
	tokens = append(tokens, this.tablename.Query())

	// Return the query
	return strings.Join(tokens, " ")
}

////////////////////////////////////////////////////////////////////////////////
// INSERT OR REPLACE

func (this *insert) WithSchema(schema string) sqlite.SQInsert {
	this.tablename.WithSchema(schema)
	return this
}

func (this *insert) DefaultValues() sqlite.SQInsert {
	this.defaultvalues = true
	return this
}

func (this *insert) Query() string {
	tokens := []string{this.class, "INTO"}

	// Add table name
	tokens = append(tokens, this.tablename.Query())

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

func (this *column) Query() string {
	if this.nullable {
		return fmt.Sprintf("%v %v", QuoteIdentifier(this.name), this.decltype)
	} else {
		return fmt.Sprintf("%v %v NOT NULL", QuoteIdentifier(this.name), this.decltype)
	}
}

func (this *column) String() string {
	str := "<sqlite.column"
	str += fmt.Sprintf(" name=%q", this.name)
	str += fmt.Sprintf(" type=%q", this.decltype)
	if !this.nullable {
		str += " notnull"
	}
	if this.primary {
		str += " primary"
	}
	return str + ">"
}
