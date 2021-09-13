package sqlite

import "strings"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SQFlag uint
type SQKey uint
type SQWriteHook func(SQResults, interface{}) error

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// SQObjects is an sqlite connection but adds ability to read, write and delete
type SQObjects interface {
	SQConnection

	// Create classes with named database and modification flags
	Create(string, SQFlag, ...SQClass) error

	// Write objects to database
	Write(v ...interface{}) ([]SQResults, error)

	// Read objects from database
	Read(SQClass) (SQIterator, error)

	// Write objects to database, call hook after each write
	WriteWithHook(SQWriteHook, ...interface{}) ([]SQResults, error)

	// Delete objects from the database
	Delete(v ...interface{}) ([]SQResults, error)
}

// SQClass is a class definition, which can be a table or view
type SQClass interface {
	// Create class in the named database with modification flags
	Create(SQConnection, string, SQFlag) error

	// Read all objects from the class and return the iterator
	// TODO: Need sort, filter, limit, offset
	Read(SQConnection) (SQIterator, error)

	// Insert objects, return rowids
	Insert(SQConnection, ...interface{}) ([]SQResults, error)

	// Update objects by primary key, return rowids
	Update(SQConnection, ...interface{}) ([]SQResults, error)

	// Upsert objects by primary key, return rowids
	Upsert(SQConnection, ...interface{}) ([]SQResults, error)

	// Delete objects from the database by primary key
	Delete(SQConnection, ...interface{}) ([]SQResults, error)

	// Set a foreign key reference to parent class and columns. Panic
	// on error, and return same class
	ForeignKey(SQClass, ...string) SQClass
}

// SQIterator is an iterator for a Read operation
type SQIterator interface {
	// Next returns the next object in the iterator, or nil if there are no more
	Next() interface{}

	// RowId returns the last read row, should be called after Next()
	RowId() int64

	// Close releases any resources associated with the iterator
	Close() error
}

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// Create flags
	SQLITE_FLAG_DELETEIFEXISTS SQFlag = 1 << iota // Delete existing database objects if they already exist
	SQLITE_FLAG_UPDATEONINSERT                    // Update existing object if a unique constraint fails

	// Other constants
	SQLITE_FLAG_NONE SQFlag = 0
	SQLITE_FLAG_MIN         = SQLITE_FLAG_DELETEIFEXISTS
	SQLITE_FLAG_MAX         = SQLITE_FLAG_UPDATEONINSERT
)

const (
	SQKeyNone SQKey = iota
	SQKeyInsert
	SQKeySelect
	SQKeyWrite
	SQKeyDelete
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
