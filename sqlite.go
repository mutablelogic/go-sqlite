package sqlite

import "context"

const (
	// TagName defines the tag name used for struct tags
	TagName = "sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// SQPool is an sqlite connection pool
type SQPool interface {
	// Close waits for all connections to be released and then releases resources
	Close() error

	// Get a connection from the pool, and return it to the pool when the context
	// is cancelled or it is put back using the Put method. If there are no
	// connections available or an error occurs, nil is returned.
	Get(context.Context) SQConnection

	// Cur returns the current number of used connections
	Cur() int64

	// Return connection to the pool
	Put(SQConnection)

	// Max returns the maximum number of connections allowed
	Max() int64

	// SetMax allowed connections released from pool. Note this does not change
	// the maximum instantly, it will settle to this value over time. Set as value
	// zero to disable opening new connections
	SetMax(int64)
}

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
	Query(SQStatement, ...interface{}) (SQResult, error)
}

// SQResult increments over returned rows from a query
type SQResult interface {
	// Return next row, returns nil when all rows consumed
	Next() []interface{}

	// Return next map of values, or nil if no more rows
	NextMap() map[string]interface{}

	// NextQuery executes the next query or returns io.EOF
	NextQuery(...interface{}) error

	// Close the rows, and free up any resources
	Close() error

	// Return Last RowID inserted of last statement
	LastInsertId() int64

	// Return number of changes made of last statement
	RowsAffected() uint64
}
