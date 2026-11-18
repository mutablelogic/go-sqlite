package ncruces

import (
	"context"
	"errors"
	"fmt"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	sqlite3 "github.com/ncruces/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// error returns SQLite's error as a DatabaseError, of the kinds of error it
// is, wrapped by the context's error, when it was interrupted, or stopped
// waiting for a lock, as it was done, or the error as it is, when it isn't
// SQLite's
func (p *pool) error(ctx context.Context, err error) error {
	var kind sqlite.Err
	var dbErr *sqlite.DatabaseError
	var sqliteErr *sqlite3.Error
	var code sqlite3.ErrorCode
	switch {
	case err == nil:
		return nil
	case errors.As(err, &kind), errors.As(err, &dbErr):
		return err
	case errors.As(err, &sqliteErr):
		err = sqlite.NewDatabaseError(err, kinds(sqliteErr.Code(), sqliteErr.ExtendedCode())...)
	case errors.As(err, &code):
		err = sqlite.NewDatabaseError(err, kinds(code, 0)...)
	default:
		return err
	}
	if ctxErr := ctx.Err(); ctxErr != nil && (errors.Is(err, sqlite3.INTERRUPT) || errors.Is(err, sqlite3.BUSY)) {
		return fmt.Errorf("%w: %w", ctxErr, err)
	}
	return err
}

// kinds returns the kinds of error of SQLite's codes
func kinds(code sqlite3.ErrorCode, extended sqlite3.ExtendedErrorCode) []sqlite.Err {
	switch extended {
	case sqlite3.CONSTRAINT_UNIQUE, sqlite3.CONSTRAINT_PRIMARYKEY:
		return []sqlite.Err{sqlite.ErrConflict, sqlite.ErrUniqueViolation}
	case sqlite3.CONSTRAINT_FOREIGNKEY:
		return []sqlite.Err{sqlite.ErrBadParameter, sqlite.ErrForeignKeyViolation}
	case sqlite3.CONSTRAINT_NOTNULL:
		return []sqlite.Err{sqlite.ErrBadParameter, sqlite.ErrNotNullViolation}
	case sqlite3.CONSTRAINT_CHECK:
		return []sqlite.Err{sqlite.ErrBadParameter, sqlite.ErrCheckViolation}
	}
	switch code {
	case sqlite3.CONSTRAINT, sqlite3.READONLY, sqlite3.MISMATCH, sqlite3.RANGE:
		return []sqlite.Err{sqlite.ErrBadParameter}
	case sqlite3.BUSY, sqlite3.LOCKED, sqlite3.CANTOPEN, sqlite3.NOTADB:
		return []sqlite.Err{sqlite.ErrNotAvailable}
	}
	return nil
}
