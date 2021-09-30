package sqlite3

import (
	// Modules
	"github.com/mutablelogic/go-sqlite/sys/sqlite3"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// DefaultFlags are the default flags for a new database connection
	DefaultFlags  = SQFlag(sqlite3.SQLITE_OPEN_CREATE | sqlite3.SQLITE_OPEN_READWRITE)
	DefaultSchema = sqlite3.DefaultSchema
	defaultMemory = sqlite3.DefaultMemory
	tempSchema    = "temp"
)

////////////////////////////////////////////////////////////////////////////////
// METHODS

func Version() string {
	str, _, _ := sqlite3.Version()
	return str
}
