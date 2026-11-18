package sqlite

import (
	"database/sql"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Driver opens SQLite databases, as a database/sql database, and says what
// their errors are, so the driver, such as ncruces's, can be replaced.
type Driver interface {
	// Open opens the database of the path, or ":memory:", with the pragmas,
	// which are set on each of its connections. It writes, beginning its
	// transactions immediately, or, when readonly, only reads, beginning
	// them when they read.
	Open(path string, readonly bool, pragmas []Pragma) (*sql.DB, error)

	// Error returns the driver's error as a DatabaseError, of the kinds of
	// error it is, such as ErrUniqueViolation, or the error as it is, when
	// it isn't the driver's
	Error(error) error
}

// Pragma is a pragma, which is set on each connection, such as
// busy_timeout, of 10000.
type Pragma struct {
	Name  string
	Value string
}
