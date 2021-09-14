package sqlite3

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	// Packages
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"
	// Namespace imports
	//. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// PoolConfig is the starting configuration for a pool
type Results struct {
	st      *sqlite3.StatementEx
	results *sqlite3.Results
	n       uint // next statement to execute
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewResults(st *sqlite3.StatementEx) *Results {
	r := new(Results)
	r.st, r.n = st, 0
	return r
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r *Results) String() string {
	str := "<results"
	if n := r.st.Count(); n > 0 {
		str += fmt.Sprint(" cached (", n, ")")
	}
	str += fmt.Sprint(" ", r.st)
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// NextQuery will execute the next query in the statement, return io.EOF if there
// are no more statements. In order to read the rows, repeatedly read the rows
// using the Next function.
func (r *Results) NextQuery(v ...interface{}) error {
	if results, err := r.st.Exec(r.n, v...); errors.Is(err, sqlite3.SQLITE_DONE) {
		return io.EOF
	} else if err != nil {
		return err
	} else {
		r.results = results
		r.n++
		// TODO: Set Columns, etc.
		return nil
	}
}

// Return a row from the results, or return io.EOF if all results have been consumed
func (r *Results) Next(t ...reflect.Type) ([]interface{}, error) {
	if r.results == nil {
		return nil, io.EOF
	} else {
		return r.results.Next(t...)
	}
}

func (r *Results) ExpandedSQL() string {
	if r.results == nil {
		return ""
	} else {
		return r.results.ExpandedSQL()
	}
}

// Return LastInsertId by last execute or -1 if no valid results
func (r *Results) LastInsertId() int64 {
	if r.results == nil {
		return -1
	} else {
		return r.results.LastInsertId()
	}
}

// Return RowsAffected by last execute or -1 if no valid results
func (r *Results) RowsAffected() int {
	if r.results == nil {
		return -1
	} else {
		return r.results.RowsAffected()
	}
}
