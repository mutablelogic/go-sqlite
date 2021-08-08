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

	// Create transaction block
	Do(func(SQTransaction) error) error

	// Close
	Close() error
}

// SQLanguage defines the interface for SQLite language
type SQLanguage interface {
	// Q creates a statement from a string
	Q(string) SQStatement

	// CreateTable creates a table with name and specified columns
	CreateTable(string, ...SQColumn) SQTable

	// CreateIndex with a source table name and defined column names
	CreateIndex(string, ...string) SQIndex

	// Column with name and declared type
	Column(string, string) SQColumn

	// DropTable with name
	DropTable(string) SQDrop

	// DropIndex with name
	DropIndex(string) SQDrop

	// DropTrigger with name
	DropTrigger(string) SQDrop

	// DropView with name
	DropView(string) SQDrop

	// Insert values into a table with a name and defined column names
	Insert(string, ...string) SQInsert

	// Replace values into a table with a name and defined column names
	Replace(string, ...string) SQInsert
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

// SQTable defines a table of columns and indexes
type SQTable interface {
	SQStatement

	IfNotExists() SQTable
	WithSchema(string) SQTable
	WithTemporary() SQTable
	WithoutRowID() SQTable
	WithIndex(...string) SQTable
	WithUnique(...string) SQTable
}

type SQInsert interface {
	SQStatement

	WithSchema(string) SQInsert
	DefaultValues() SQInsert
}

type SQSelect interface {
	SQStatement

	WithDistinct() SQSelect
}

type SQIndex interface {
	SQStatement

	IfNotExists() SQIndex
	WithName(string) SQIndex
	WithSchema(string) SQIndex
	WithUnique() SQIndex
}

type SQDrop interface {
	SQStatement

	IfExists() SQDrop
	WithSchema(string) SQDrop
}

type SQColumn interface {
	Primary() SQColumn
	NotNull() SQColumn
}

type SQSource interface {
	WithAlias(string) SQSource
}
