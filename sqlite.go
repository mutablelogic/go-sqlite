package sqlite

///////////////////////////////////////////////////////////////////////////////
// INTERFACES - CONNECTION

// SQConnection is an sqlite connection to one or more databases
type SQConnection interface {
	SQTransaction
	SQLanguage

	// Schemas returns a list of all the schemas in the database
	Schemas() []string

	// Table returns a list of non-temporary tables in the default schema
	Tables() []string

	// TablesEx returns a list of tables in the specified schema. Pass true
	// as second argument to include temporary tables.
	TablesEx(string, bool) []string

	// ColumnsEx returns an ordered list of columns in the specified table
	Columns(string) []SQColumn

	// ColumnsEx returns an ordered list of columns in the specified table and schema
	ColumnsEx(string, string) []SQColumn

	// Attach with schema name to a database at path in second argument
	Attach(string, string) error

	// Detach database by schema name
	Detach(string) error

	// Create transaction block, rollback on error
	Do(func(SQTransaction) error) error

	// Close
	Close() error
}

// SQLanguage defines the interface for SQLite language
type SQLanguage interface {
	// Q creates an statement which can be used in Exec or Query
	Q(interface{}) SQStatement

	// N creates a name (table or column name)
	N(string) SQName

	// P creates a bound parameter
	P() SQExpr

	// V creates a value
	V(interface{}) SQExpr

	// S selects data from one or more data sources
	S(...SQName) SQSelect
}

// SQTransaction is an sqlite transaction
type SQTransaction interface {
	// Query and return a set of results
	Query(SQStatement, ...interface{}) (SQRows, error)

	// Execute a statement and return affected rows
	Exec(SQStatement, ...interface{}) (SQResult, error)
}

// SQRows increments over returned rows from a query
type SQRows interface {
	// Return next row, returns io.EOF when all rows consumed
	Next(v interface{}) error

	// Return next map of values, or nil if no more rows
	NextMap() map[string]interface{}

	// Return next array of values, or nil if no more rows
	NextArray() []interface{}

	// Close the rows, and free up any resources
	Close() error
}

// SQResult is returned after SQTransaction.Exec
type SQResult struct {
	LastInsertId int64
	RowsAffected uint64
}

// SQStatement is any statement which can be executed
type SQStatement interface {
	Query() string
}

// SQName defines a table or column name
type SQName interface {
	SQStatement
	SQExpr

	// Modify the source
	WithSchema(string) SQName
	WithType(string) SQColumn
	WithAlias(string) SQName

	// Insert or replace a row with named columns
	Insert(...string) SQInsert
	Replace(...string) SQInsert

	// Create objects
	CreateTable(...SQColumn) SQTable
	CreateView(SQSelect, ...string) SQIndexView
	//CreateIndex(SQName, ...SQColumn) SQIndexView

	// Drop objects
	DropTable() SQDrop
	DropIndex() SQDrop
	DropTrigger() SQDrop
	DropView() SQDrop
}

// SQTable defines a table of columns and indexes
type SQTable interface {
	SQStatement

	IfNotExists() SQTable
	WithTemporary() SQTable
	WithoutRowID() SQTable
	WithIndex(...string) SQTable
	WithUnique(...string) SQTable
}

// SQIndexView defines a create index or view statement
type SQIndexView interface {
	SQStatement

	IfNotExists() SQIndexView
	WithTemporary() SQIndexView
	WithUnique() SQIndexView
}

// SQDrop defines a drop for tables, views, indexes, and triggers
type SQDrop interface {
	SQStatement

	IfExists() SQDrop
}

// SQInsert defines an insert or replace statement
type SQInsert interface {
	SQStatement

	DefaultValues() SQInsert
}

// SQSelect defines a select statement
type SQSelect interface {
	SQStatement

	// Set select flags
	WithDistinct() SQSelect
	WithLimitOffset(limit, offset uint) SQSelect
	Where(...interface{}) SQSelect
}

// SQAlter defines an alter table statement
type SQAlter interface {
	SQStatement

	WithSchema(string) SQAlter
}

// SQColumn represents a column definition
type SQColumn interface {
	Primary() SQColumn
	NotNull() SQColumn
}

// SQExpr defines any expression
type SQExpr interface {
	SQStatement

	// Comparison expression with one or more right hand side expressions
	Is(SQExpr, ...SQExpr) SQComparison
}

// SQComparison defines a comparison between two expressions
type SQComparison interface {
	SQStatement

	// Negate the comparison
	Not() SQComparison
}

/*
	Gt() SQComparison
	GtEq() SQComparison
	Lt() SQComparison
	LtEq() SQComparison
*/
