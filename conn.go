// Package sqlite is a SQLite client, of the same shape as go-pg's: queries
// are bound to a connection, and objects are read and written by the
// Reader, Writer and Selector interfaces. An implementation, such as
// package ncruces, opens a database as a PoolConn.
package sqlite

import (
	"context"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// PoolConn is a database, opened from a file or in memory, which writes on
// one connection, and reads on many.
type PoolConn interface {
	Conn

	// Ping the database
	Ping(context.Context) error

	// Close the database
	Close() error
}

// Conn is a connection, with bound parameters and queries.
type Conn interface {
	// Return a new connection with bound parameters
	With(...any) Conn

	// Return a new connection with bound queries
	WithQueries(...*Queries) Conn

	// Perform a transaction within a function, on the connection which
	// writes, which is begun immediately, so it isn't upgraded to write
	// later, when another transaction may be writing
	Tx(context.Context, func(Conn) error) error

	// Perform a bulk operation within a function, in a transaction, whose
	// statements are prepared once
	Bulk(context.Context, func(Conn) error) error

	// Execute a query
	Exec(context.Context, string) error

	// Perform an insert
	Insert(context.Context, Reader, Writer) error

	// Perform an update
	Update(context.Context, Reader, Selector, Writer) error

	// Perform a delete
	Delete(context.Context, Reader, Selector) error

	// Perform a get, on a connection which reads, unless in a transaction
	Get(context.Context, Reader, Selector) error

	// Perform a list, on a connection which reads, unless in a transaction.
	// If the reader is a ListReader, then the count of items is also
	// calculated
	List(context.Context, Reader, Selector) error
}

// Op represents a database operation type.
type Op uint

// Row is a row of a query's results, which is scanned into values.
type Row interface {
	Scan(...any) error
}

// Reader scans a database row into an object.
type Reader interface {
	// Scan row into a result
	Scan(Row) error
}

// ListReader scans database rows and counts total results.
type ListReader interface {
	Reader

	// Scan count into the result
	ScanCount(Row) error
}

// Writer binds object fields for insert or update operations.
type Writer interface {
	// Set bind parameters for an insert
	Insert(*Bind) (string, error)

	// Set bind parameters for an update
	Update(*Bind) error
}

// Selector binds parameters for get, update, or delete operations.
type Selector interface {
	// Set bind parameters for getting, updating or deleting
	Select(*Bind, Op) (string, error)
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

// Operations
const (
	None Op = iota
	Get
	Insert
	Update
	Delete
	List
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o Op) String() string {
	switch o {
	case Get:
		return "GET"
	case Insert:
		return "INSERT"
	case Update:
		return "UPDATE"
	case Delete:
		return "DELETE"
	case List:
		return "LIST"
	}
	return "UNKNOWN"
}
