package sqlite3

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	// Packages
	sqlite3 "github.com/mutablelogic/go-sqlite/sys/sqlite3"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
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

// Return the columns for the current results
func (r *Results) Columns() []SQColumn {
	if r.results == nil {
		return nil
	}
	cols := make([]SQColumn, r.results.ColumnCount())
	for i := range cols {
		cols[i] = C(r.results.ColumnName(i)).WithType(r.results.ColumnDeclType(i))
	}
	return cols
}

func (r *Results) ColumnSource(i int) (string, string, string) {
	if r.results == nil {
		return "", "", ""
	}
	schema, table, name := r.results.ColumnDatabaseName(i), r.results.ColumnTableName(i), r.results.ColumnName(i)
	if name == "" {
		name = r.results.ColumnOriginName(i)
	}
	return schema, table, name
}
