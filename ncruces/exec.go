package ncruces

import (
	"fmt"
	"strings"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	values "github.com/mutablelogic/go-sqlite/values"
	sqlite3 "github.com/ncruces/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// row is a statement's row, which is scanned into values
type row struct {
	stmt *sqlite3.Stmt
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

////////////////////////////////////////////////////////////////////////////////
// EXECUTE

// execute executes the query's statements, with its ${key} replaced, and its
// @key arguments bound, and scans their rows into the reader, which is
// ErrNotFound without any
func execute(db *sqlite3.Conn, bind *sqlite.Bind, query string, reader sqlite.Reader) error {
	return executeRows(db, bind, query, reader, false)
}

// executeOne executes the query's statements, as execute does, and scans the
// first row into the reader, without reading the others
func executeOne(db *sqlite3.Conn, bind *sqlite.Bind, query string, reader sqlite.Reader) error {
	return executeRows(db, bind, query, reader, true)
}

func executeRows(db *sqlite3.Conn, bind *sqlite.Bind, query string, reader sqlite.Reader, one bool) error {
	query, vars := bind.Expand(query)
	scanned := false
	for strings.TrimSpace(query) != "" {
		stmt, tail, err := db.Prepare(query)
		if err != nil {
			return err
		}
		query = tail
		if stmt == nil {
			// A comment, or a ;
			continue
		}
		if err := step(stmt, vars, reader, one, &scanned); err != nil {
			return err
		}
	}
	if reader != nil && !scanned {
		return sqlite.ErrNotFound
	}
	return nil
}

// step binds the statement's arguments, and scans its rows into the reader
func step(stmt *sqlite3.Stmt, vars map[string]any, reader sqlite.Reader, one bool, scanned *bool) error {
	defer stmt.Close()
	for i := 1; i <= stmt.BindCount(); i++ {
		name := stmt.BindName(i)
		if name == "" {
			// A positional argument, which is NULL
			continue
		}
		if err := values.Set(param{stmt, i}, vars[name[1:]]); err != nil {
			return sqlite.ErrBadParameter.Withf("%s: %w", name, err)
		}
	}
	for stmt.Step() {
		if reader == nil || (one && *scanned) {
			continue
		}
		if err := reader.Scan(row{stmt}); err != nil {
			return err
		}
		*scanned = true
		if one {
			// The statement is done, without reading its other rows
			return nil
		}
	}
	return stmt.Err()
}

////////////////////////////////////////////////////////////////////////////////
// SCAN

// Scan scans the row's columns into the values, which are pointers, of each
// of them, as values.Scan does.
func (r row) Scan(dest ...any) error {
	if n := r.stmt.ColumnCount(); len(dest) != n {
		return fmt.Errorf("%d values, for %d columns", len(dest), n)
	}
	for i, dest := range dest {
		if err := values.Scan(column{r.stmt, i}, dest); err != nil {
			return fmt.Errorf("column %q: %w", r.stmt.ColumnName(i), err)
		}
	}
	return nil
}
