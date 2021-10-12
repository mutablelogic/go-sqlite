package sqlite

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// SQStatement is any statement which can be prepared or executed
type SQStatement interface {
	SQExpr
	Query() string
}

// SQSource defines a table or column name
type SQSource interface {
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
	CreateTrigger(string, ...SQStatement) SQTrigger
	CreateView(SQSelect, ...string) SQIndexView
	ForeignKey(...string) SQForeignKey

	// Alter objects
	AlterTable() SQAlter

	// Update and delete data
	Update(...string) SQUpdate
	Delete(...interface{}) SQStatement
}

// SQJoin defines one or more joins
type SQJoin interface {
	SQExpr

	Join(...SQExpr) SQJoin
	LeftJoin(...SQExpr) SQJoin
	LeftInnerJoin(...SQExpr) SQJoin
	Using(...string) SQJoin
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
	Options(...string) SQIndexView
}

// SQTrigger defines a create trigger statement
type SQTrigger interface {
	SQStatement

	// Modifiers
	IfNotExists() SQTrigger
	WithTemporary() SQTrigger
	Before() SQTrigger
	After() SQTrigger
	InsteadOf() SQTrigger
	Delete() SQTrigger
	Insert() SQTrigger
	Update(...string) SQTrigger
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

	// Destination expressions for results
	To(...SQExpr) SQSelect

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
	SQExpr

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
	String() string
}
