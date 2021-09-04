package sqlite

import "strings"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SQFlag uint
type SQKey uint
type SQWriteHook func(SQResult, interface{}) error

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// SQObjects is an sqlite connection but adds ability to read, write and delete
type SQObjects interface {
	SQConnection

	// Register a new named table based on prototype
	Register(string, interface{}) (SQClass, error)

	// Create schema for a class (and drop existing data as necessary)
	Create(SQClass, SQFlag) error

	// Write objects to database
	Write(v ...interface{}) ([]SQResult, error)

	// Read objects from database
	Read(SQClass) (SQIterator, error)

	// Write objects to database, call hook after each write
	WriteWithHook(SQWriteHook, ...interface{}) ([]SQResult, error)

	// Delete objects from the database
	Delete(v ...interface{}) ([]SQResult, error)
}

// SQClass is a class definition
type SQClass interface {
	SQSource

	// Set a foreign key reference to class
	WithForeignKey(SQClass, ...string) error

	// Return a prototype object for this class
	Proto() interface{}

	// Get statements related to a key
	Get(SQKey) []SQStatement

	// Set statements, return a new key
	Set(...SQStatement) SQKey
}

// SQIterator is an iterator for a Read operation
type SQIterator interface {
	// Next returns the next object in the iterator, or nil if there are no more
	Next() interface{}

	// Close releases any resources associated with the iterator
	Close() error
}

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// Delete existing database objects if they already exist
	SQLITE_FLAG_DELETEIFEXISTS SQFlag = 1 << iota

	// Other constants
	SQLITE_FLAG_NONE SQFlag = 0
	SQLITE_FLAG_MIN         = SQLITE_FLAG_DELETEIFEXISTS
	SQLITE_FLAG_MAX         = SQLITE_FLAG_DELETEIFEXISTS
)

const (
	SQKeyNone SQKey = iota
	SQKeyCreate
	SQKeyWrite
	SQKeyDelete
	SQKeyRead
	SQKeyGetRowId
	SQKeyMax
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f SQFlag) String() string {
	if f == SQLITE_FLAG_NONE {
		return f.FlagString()
	}
	str := ""
	for v := SQLITE_FLAG_MIN; v <= SQLITE_FLAG_MAX; v <<= 1 {
		if f&v == v {
			str += v.FlagString() + "|"
		}
	}
	return strings.TrimSuffix(str, "|")
}

func (v SQFlag) FlagString() string {
	switch v {
	case SQLITE_FLAG_NONE:
		return "SQLITE_FLAG_NONE"
	case SQLITE_FLAG_DELETEIFEXISTS:
		return "SQLITE_FLAG_DELETEIFEXISTS"
	default:
		return "[?? Invalid SQFlag]"
	}
}
