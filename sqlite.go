package sqlite

import (
	"context"
	"strings"
)

const (
	// TagName defines the tag name used for struct tags
	TagName = "sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	SQAuthFlag uint
	SQTxnFlag  uint
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
	//SQTransaction

	// Execute a transaction with context, rollback on any errors
	// or cancelled context
	Do(context.Context, SQTxnFlag, func(SQTransaction) error) error

	// Schemas returns a list of all the schemas in the database
	Schemas() []string

	// Tables returns a list of tables in a schema
	Tables(string) []string

	// Filename returns a filename for a schema, returns empty
	// string if in-memory database
	Filename(string) string

	// ColumnsForTable returns the columns in a schema and table
	ColumnsForTable(string, string) []SQColumn

	// ColumnsForIndex returns the column names associated with schema and index
	ColumnsForIndex(string, string) []string

	// IndexesForTable returns the indexes associated with a schema and table
	IndexesForTable(string, string) []SQIndexView

	// Views returns a list of view names in a schema
	Views(string) []string

	// Modules returns a list of modules. If an argument is
	// provided, then only modules with those name prefixes
	// matched
	Modules(...string) []string
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

// SQAuth is an interface for authenticating an action
type SQAuth interface {
	// CanSelect is called to authenticate a SELECT
	CanSelect(context.Context) error

	// CanTransaction is called for BEGIN, COMMIT, or ROLLBACK
	CanTransaction(context.Context, SQAuthFlag) error

	// CanExec is called to authenticate an operation other then SELECT
	CanExec(context.Context, SQAuthFlag, string, ...string) error
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	SQLITE_TXN_DEFAULT SQTxnFlag = (1 << iota)
	SQLITE_TXN_IMMEDIATE
	SQLITE_TXN_EXCLUSIVE
	SQLITE_TXN_SNAPSHOT
	SQLITE_TXN_NO_FOREIGNKEY_CONSTRAINTS
)

const (
	SQLITE_AUTH_TABLE       SQAuthFlag = 1 << iota // Table Object
	SQLITE_AUTH_INDEX                              // Index Object
	SQLITE_AUTH_VIEW                               // View Object
	SQLITE_AUTH_TRIGGER                            // Trigger Object
	SQLITE_AUTH_VTABLE                             // Virtual Table Object
	SQLITE_AUTH_TEMP                               // Temporary Object
	SQLITE_AUTH_TRANSACTION                        // Transaction
	SQLITE_AUTH_CREATE                             // Create operation
	SQLITE_AUTH_DROP                               // Drop operation
	SQLITE_AUTH_INSERT                             // Insert operation
	SQLITE_AUTH_DELETE                             // Delete operation
	SQLITE_AUTH_ALTER                              // Alter operation
	SQLITE_AUTH_ANALYZE                            // Analyze  operation
	SQLITE_AUTH_PRAGMA                             // Pragma operation
	SQLITE_AUTH_READ                               // Read column operation
	SQLITE_AUTH_UPDATE                             // Update column operation
	SQLITE_AUTH_FUNCTION                           // Execute function operation
	SQLITE_AUTH_BEGIN                              // Begin txn operation
	SQLITE_AUTH_COMMIT                             // Commit txn operation
	SQLITE_AUTH_ROLLBACK                           // Rollback txn operation
	SQLITE_AUTH_MIN                    = SQLITE_AUTH_TABLE
	SQLITE_AUTH_MAX                    = SQLITE_AUTH_ROLLBACK
	SQLITE_AUTH_NONE        SQAuthFlag = 0
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v SQAuthFlag) StringFlag() string {
	switch v {
	case SQLITE_AUTH_NONE:
		return "SQLITE_AUTH_NONE"
	case SQLITE_AUTH_TABLE:
		return "SQLITE_AUTH_TABLE"
	case SQLITE_AUTH_INDEX:
		return "SQLITE_AUTH_INDEX"
	case SQLITE_AUTH_VIEW:
		return "SQLITE_AUTH_VIEW"
	case SQLITE_AUTH_TRIGGER:
		return "SQLITE_AUTH_TRIGGER"
	case SQLITE_AUTH_VTABLE:
		return "SQLITE_AUTH_VTABLE"
	case SQLITE_AUTH_TEMP:
		return "SQLITE_AUTH_TEMP"
	case SQLITE_AUTH_TRANSACTION:
		return "SQLITE_AUTH_TRANSACTION"
	case SQLITE_AUTH_CREATE:
		return "SQLITE_AUTH_CREATE"
	case SQLITE_AUTH_DROP:
		return "SQLITE_AUTH_DROP"
	case SQLITE_AUTH_INSERT:
		return "SQLITE_AUTH_INSERT"
	case SQLITE_AUTH_DELETE:
		return "SQLITE_AUTH_DELETE"
	case SQLITE_AUTH_ALTER:
		return "SQLITE_AUTH_ALTER"
	case SQLITE_AUTH_ANALYZE:
		return "SQLITE_AUTH_ANALYZE"
	case SQLITE_AUTH_READ:
		return "SQLITE_AUTH_READ"
	case SQLITE_AUTH_UPDATE:
		return "SQLITE_AUTH_UPDATE"
	case SQLITE_AUTH_PRAGMA:
		return "SQLITE_AUTH_PRAGMA"
	case SQLITE_AUTH_FUNCTION:
		return "SQLITE_AUTH_FUNCTION"
	case SQLITE_AUTH_BEGIN:
		return "SQLITE_AUTH_BEGIN"
	case SQLITE_AUTH_COMMIT:
		return "SQLITE_AUTH_COMMIT"
	case SQLITE_AUTH_ROLLBACK:
		return "SQLITE_AUTH_ROLLBACK"
	default:
		return "[?? Invalid SQAuthFlag value]"
	}
}

func (v SQAuthFlag) String() string {
	if v == SQLITE_AUTH_NONE {
		return v.StringFlag()
	}
	str := ""
	for f := SQLITE_AUTH_MIN; f <= SQLITE_AUTH_MAX; f = f << 1 {
		if v&f == f {
			str += "|" + f.StringFlag()
		}
	}
	return strings.TrimPrefix(str, "|")
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Is any of the flags in q
func (v SQAuthFlag) Is(q SQAuthFlag) bool {
	return v&q != 0
}

// Is any of the flags in q
func (v SQTxnFlag) Is(q SQTxnFlag) bool {
	return v&q != 0
}
