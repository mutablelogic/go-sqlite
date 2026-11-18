package ncruces

import (
	"errors"
	"fmt"
	"strings"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	values "github.com/mutablelogic/go-sqlite/values"
	sqlite3 "github.com/ncruces/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// statements are a bulk transaction's prepared statements, by their SQL,
// with its ${key} replaced, which are reused, rather than prepared again
type statements map[string]*sqlite3.Stmt

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// maxStatements is the most statements a bulk transaction reuses, as a
	// query whose ${key} differs, such as of a value, rather than an @key
	// argument, is another statement; others are prepared, and closed
	maxStatements = 64
)

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
func execute(db *sqlite3.Conn, cache statements, bind *sqlite.Bind, query string, reader sqlite.Reader) error {
	return executeRows(db, cache, bind, query, reader, false)
}

// executeOne executes the query's statements, as execute does, and scans the
// first row into the reader, without reading the others
func executeOne(db *sqlite3.Conn, cache statements, bind *sqlite.Bind, query string, reader sqlite.Reader) error {
	return executeRows(db, cache, bind, query, reader, true)
}

// executeRows executes the query's statements, which are reused, of the
// cache, when it isn't nil, and the query is one statement, or else
// prepared, and closed
func executeRows(db *sqlite3.Conn, cache statements, bind *sqlite.Bind, query string, reader sqlite.Reader, one bool) error {
	query, vars := bind.Expand(query)
	scanned := false
	for strings.TrimSpace(query) != "" {
		stmt, cached := cache[query]
		if cached {
			query = ""
		} else {
			var tail string
			var err error
			if stmt, tail, err = db.Prepare(query); err != nil {
				return err
			}
			if stmt == nil {
				// A comment, or a ;
				query = tail
				continue
			}
			if cache != nil && len(cache) < maxStatements && strings.TrimSpace(tail) == "" {
				cache[query] = stmt
				cached = true
			}
			query = tail
		}
		err := step(stmt, vars, reader, one, &scanned)
		if cached {
			// It's reset, so it doesn't read, or lock, until it's reused
			_ = stmt.Reset()
			err = errors.Join(err, stmt.ClearBindings())
		} else {
			_ = stmt.Close()
		}
		if err != nil {
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

////////////////////////////////////////////////////////////////////////////////
// STATEMENTS

// close closes the statements
func (s statements) close() error {
	var result error
	for query, stmt := range s {
		result = errors.Join(result, stmt.Close())
		delete(s, query)
	}
	return result
}
