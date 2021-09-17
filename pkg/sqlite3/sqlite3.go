package sqlite3

import (
	// Modules
	"github.com/djthorpe/go-sqlite/sys/sqlite3"

	// Namespace imports
	. "github.com/djthorpe/go-sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// DefaultFlags are the default flags for a new database connection
	DefaultFlags  = SQFlag(sqlite3.SQLITE_OPEN_CREATE | sqlite3.SQLITE_OPEN_READWRITE)
	defaultMemory = sqlite3.DefaultMemory
	defaultSchema = sqlite3.DefaultSchema
	tempSchema    = "temp"
)

////////////////////////////////////////////////////////////////////////////////
// METHODS

func Version() string {
	str, _, _ := sqlite3.Version()
	return str
}
