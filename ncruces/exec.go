package ncruces

import (
	"context"
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
	// multiSavepoint is the savepoint of a query of several statements
	multiSavepoint = "statements"

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

// executeRows executes the query's statements, whose ${key} are replaced,
// with their arguments bound, of the vars, which are set in args, when it
// isn't nil, and scans their rows, or the first, into the reader, which is
// ErrNotFound without any. They're reused, of the cache, when it isn't nil,
// and the query is one statement, or else prepared, and closed.
func executeRows(db *sqlite3.Conn, cache statements, query string, vars map[string]any, reader sqlite.Reader, one bool, args map[string]any) (err error) {
	// A query of several statements is executed in a savepoint, so all of
	// them are, or, when one fails, or a reader panics, none are
	var savepoint, done bool
	defer func() {
		if !savepoint {
			return
		}
		if err == nil && done {
			err = db.Exec("RELEASE " + multiSavepoint)
			return
		}
		old := db.SetInterrupt(context.Background())
		defer db.SetInterrupt(old)
		err = errors.Join(err, db.Exec("ROLLBACK TO "+multiSavepoint), db.Exec("RELEASE "+multiSavepoint))
	}()

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
			if !savepoint && strings.TrimSpace(tail) != "" {
				if err := db.Exec("SAVEPOINT " + multiSavepoint); err != nil {
					_ = stmt.Close()
					return err
				}
				savepoint = true
			}
			query = tail
		}
		if err := run(stmt, cached, vars, args, reader, one, &scanned); err != nil {
			return err
		}
	}
	done = true
	if reader != nil && !scanned {
		return sqlite.ErrNotFound
	}
	return nil
}

// run steps the statement, and then resets it, when it's cached, so it
// doesn't read, or lock, until it's reused, or closes it, even when it
// panics, such as when it's scanned
func run(stmt *sqlite3.Stmt, cached bool, vars, args map[string]any, reader sqlite.Reader, one bool, scanned *bool) (err error) {
	defer func() {
		if cached {
			_ = stmt.Reset()
			err = errors.Join(err, stmt.ClearBindings())
		} else {
			_ = stmt.Close()
		}
	}()
	return step(stmt, vars, args, reader, one, scanned)
}

// step binds the statement's arguments, and scans its rows into the reader
func step(stmt *sqlite3.Stmt, vars, args map[string]any, reader sqlite.Reader, one bool, scanned *bool) error {
	for i := 1; i <= stmt.BindCount(); i++ {
		name := stmt.BindName(i)
		if name == "" {
			// A positional argument, which is NULL
			continue
		}
		if err := values.Set(param{stmt, i}, vars[name[1:]]); err != nil {
			return sqlite.ErrBadParameter.Withf("%s: %w", name, err)
		}
		if args != nil {
			args[name[1:]] = vars[name[1:]]
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
