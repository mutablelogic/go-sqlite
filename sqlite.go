package sqlite

const (
	// TagName defines the tag name used for struct tags
	TagName = "sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES - CONNECTION

// SQConnection is an sqlite connection to one or more databases
type SQConnection interface {
	SQTransaction

	// Schemas returns a list of all the schemas in the database
	Schemas() []string

	// Filename returns a filename for a schema, returns empty
	// string if in-memory database
	Filename(string) string

	// Table returns a list of non-temporary tables in the default schema
	Tables() []string

	// TablesEx returns a list of tables in the specified schema. Pass true
	// as second argument to include temporary tables.
	TablesEx(string, bool) []string

	// Indexes returns indexes for a specified table
	Indexes(string) []SQIndexView

	// Indexes returns indexes for a specified table in the specified schema
	IndexesEx(string, string) []SQIndexView

	// ColumnsEx returns an ordered list of columns in the specified table
	Columns(string) []SQColumn

	// ColumnsEx returns an ordered list of columns in the specified table and schema
	ColumnsEx(string, string) []SQColumn

	// Attach with schema name to a database at path in second argument
	Attach(string, string) error

	// Detach database by schema name
	Detach(string) error

	// Modules returns a list of virtual table modules. If a string is provided,
	// only modules with the prefix of the string will be returned.
	Modules(...string) []string

	// Create transaction block, rollback on error
	Do(func(SQTransaction) error) error

	// Get and set foreign key constraints
	ForeignKeyConstraints() (bool, error)
	SetForeignKeyConstraints(bool) error

	// Close
	Close() error
}

// SQTransaction is an sqlite transaction
type SQTransaction interface {
	// Query and return a set of results
	Query(SQStatement, ...interface{}) (SQRows, error)

	// Execute a statement and return affected rows
	Exec(SQStatement, ...interface{}) (SQResult, error)

	// Prepare a statement
	Prepare(SQStatement) (SQStatement, error)
}

// SQRows increments over returned rows from a query
type SQRows interface {
	// Return next row, returns nil when all rows consumed
	Next() []interface{}

	// Return next map of values, or nil if no more rows
	NextMap() map[string]interface{}

	// Close the rows, and free up any resources
	Close() error
}

// SQResult is returned after SQTransaction.Exec
type SQResult struct {
	LastInsertId int64
	RowsAffected uint64
}

// SQStatement is any statement which can be prepared or executed
type SQStatement interface {
	Query() string
}

// SQSource defines a table or column name
type SQSource interface {
	SQStatement
	SQExpr

	// Return name, schema, type
	Name() string
	Schema() string
	Alias() string

	// Modify the source
	WithName(string) SQSource
	WithSchema(string) SQSource
	WithType(string) SQColumn
	WithAlias(string) SQSource
	WithDesc() SQSource

	// Insert, replace or upsert a row with named columns
	Insert(...string) SQInsert
	Replace(...string) SQInsert

	// Drop objects
	DropTable() SQDrop
	DropIndex() SQDrop
	DropTrigger() SQDrop
	DropView() SQDrop

	// Create objects
	CreateTable(...SQColumn) SQTable
	CreateVirtualTable(string, ...string) SQIndexView
	CreateIndex(string, ...string) SQIndexView
	//CreateView(SQSelect, ...string) SQIndexView
	ForeignKey(...string) SQForeignKey

	// Alter objects
	AlterTable() SQAlter

	// Update and delete data
	Update(...string) SQUpdate
	Delete(...interface{}) SQStatement
}

// SQTable defines a table of columns and indexes
type SQTable interface {
	SQStatement

	IfNotExists() SQTable
	WithTemporary() SQTable
	WithoutRowID() SQTable
	WithIndex(...string) SQTable
	WithUnique(...string) SQTable
	WithForeignKey(SQForeignKey, ...string) SQTable
}

// SQUpdate defines an update statement
type SQUpdate interface {
	SQStatement

	// Modifiers
	WithAbort() SQUpdate
	WithFail() SQUpdate
	WithIgnore() SQUpdate
	WithReplace() SQUpdate
	WithRollback() SQUpdate

	// Where clause
	Where(...interface{}) SQUpdate
}

// SQIndexView defines a create index or view statement
type SQIndexView interface {
	SQStatement
	SQSource

	// Return properties
	Unique() bool
	Table() string
	Columns() []string
	Auto() bool

	// Modifiers
	IfNotExists() SQIndexView
	WithTemporary() SQIndexView
	WithUnique() SQIndexView
	WithAuto() SQIndexView
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
	WithConflictDoNothing(...string) SQInsert
	WithConflictUpdate(...string) SQInsert
}

// SQSelect defines a select statement
type SQSelect interface {
	SQStatement

	// Set select flags
	WithDistinct() SQSelect
	WithLimitOffset(limit, offset uint) SQSelect

	// Destination columns for results
	To(...SQSource) SQSelect

	// Where and order clauses
	Where(...interface{}) SQSelect
	Order(...SQSource) SQSelect
}

// SQAlter defines an alter table statement
type SQAlter interface {
	SQStatement

	// Alter operation
	AddColumn(SQColumn) SQStatement
	DropColumn(SQColumn) SQStatement
}

// SQForeignKey represents a foreign key constraint
type SQForeignKey interface {
	// Modifiers
	OnDeleteCascade() SQForeignKey
}

// SQColumn represents a column definition
type SQColumn interface {
	SQStatement

	// Properties
	Name() string
	Type() string
	Nullable() bool
	Primary() string

	// Modifiers
	NotNull() SQColumn
	WithType(string) SQColumn
	WithAlias(string) SQSource
	WithPrimary() SQColumn
	WithAutoIncrement() SQColumn
	WithDefault(v interface{}) SQColumn
	WithDefaultNow() SQColumn
}

// SQExpr defines any expression
type SQExpr interface {
	SQStatement

	// And, Or, Not
	Or(interface{}) SQExpr

	// Comparison expression with one or more right hand side expressions
	//Is(SQExpr, ...SQExpr) SQComparison
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
