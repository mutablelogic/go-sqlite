package sqlite3

import (
	// Packages
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// PoolConfig is the starting configuration for a pool
type Results struct {
	*sqlite3.Results
	n int
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewResults(v *sqlite3.Results) *Results {
	r := new(Results)
	r.Results = v
	r.n = 0
	return r
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return next row, returns nil when all rows consumed
func (r *Results) Next() []interface{} {
	if v, err := r.Results.Next(); err != nil {
		return nil
	} else {
		return v
	}
}

// Return next map of values, or nil if no more rows
func (r *Results) NextMap() map[string]interface{} {
	// TODO
	return nil
}

// NextQuery executes the next query or returns io.EOF
func (r *Results) NextQuery(...interface{}) error {
	return ErrNotImplemented
}

// Close the rows, and free up any resources
func (r *Results) Close() error {
	// TODO
	return nil
}

// Return Last RowID inserted of last statement
func (r *Results) LastInsertId() int64 {
	return -1
}

// Return number of changes made of last statement
func (r *Results) RowsAffected() int64 {
	return -1
}
