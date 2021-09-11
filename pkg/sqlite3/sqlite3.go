package sqlite3

import (
	// Modules
	"github.com/djthorpe/go-sqlite/sys/sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultSchema = sqlite3.DefaultSchema
	tempSchema    = "temp"
	defaultMemory = sqlite3.DefaultMemory
)

////////////////////////////////////////////////////////////////////////////////
// METHODS

func Version() string {
	str, _, _ := sqlite3.Version()
	return str
}

func IsComplete(v string) bool {
	return sqlite3.IsComplete(v)
}
