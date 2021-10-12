package sqlite

///////////////////////////////////////////////////////////////////////////////
// TYPES

//type SQWriteHook func(SQResults, interface{}) error

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// SQObjects is an sqlite connection but adds ability to read, write and delete
/*
type SQObjects interface {
	SQConnection

	// Create classes with named database
	Create(context.Context, string, ...SQClass) error

	// Write objects to database
	Write(v ...interface{}) ([]SQResults, error)

	// Read objects from database
	Read(SQClass) (SQIterator, error)

	// Write objects to database, call hook after each write
	WriteWithHook(SQWriteHook, ...interface{}) ([]SQResults, error)

	// Delete objects from the database
	Delete(v ...interface{}) ([]SQResults, error)
}
*/

// SQClass is a class definition, which can be a table or view
type SQClass interface {
	// Create class in the named database schema
	Create(SQTransaction, string) error

	// Read all objects from the class and return the iterator
	// TODO: Need sort, filter, limit, offset
	Read(SQTransaction) (SQIterator, error)

	// Insert objects, return rowids
	Insert(SQTransaction, ...interface{}) ([]int64, error)

	// Delete rows in table based on rowid. Returns number of deleted rows
	DeleteRows(SQTransaction, []int64) (int, error)

	// Delete keys in table based on primary keys. Returns number of deleted rows
	DeleteKeys(SQTransaction, ...interface{}) (int, error)

	// Update objects by primary key, return number of updated rows
	UpdateKeys(SQTransaction, ...interface{}) (int, error)

	// Upsert (insert or update) objects by primary key, return rowids
	UpsertKeys(SQTransaction, ...interface{}) ([]int64, error)
}

// SQIterator is an iterator for a Read operation
type SQIterator interface {
	// Next returns the next object in the iterator, or nil if there are no more
	Next() interface{}

	// RowId returns the last read row, should be called after Next()
	RowId() int64
}
