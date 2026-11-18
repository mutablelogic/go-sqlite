package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Err is a kind of error, which SQLite's errors are mapped to
type Err int

// DatabaseError is a driver's error, which is ErrDatabase, and the kinds of
// error it is, such as ErrConflict and ErrUniqueViolation, with errors.Is
type DatabaseError struct {
	err   error
	kinds []Err
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ErrSuccess Err = iota
	ErrNotFound
	ErrNotImplemented
	ErrBadParameter
	ErrNotAvailable
	ErrConflict
	ErrDatabase
	ErrUniqueViolation
	ErrForeignKeyViolation
	ErrNotNullViolation
	ErrCheckViolation
	ErrInternalServerError
)

/////////////////////////////////////////////////////////////////////
// STRINGIFY

// Error returns the string representation of the error.
func (e Err) Error() string {
	switch e {
	case ErrSuccess:
		return "success"
	case ErrNotFound:
		return "not found"
	case ErrNotImplemented:
		return "not implemented"
	case ErrBadParameter:
		return "bad parameter"
	case ErrNotAvailable:
		return "not available"
	case ErrConflict:
		return "conflict"
	case ErrDatabase:
		return "database error"
	case ErrUniqueViolation:
		return "unique violation"
	case ErrForeignKeyViolation:
		return "foreign key violation"
	case ErrNotNullViolation:
		return "not null violation"
	case ErrCheckViolation:
		return "check violation"
	case ErrInternalServerError:
		return "internal server error"
	default:
		return fmt.Sprint("Unknown error ", int(e))
	}
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// With returns the error with additional context appended.
func (e Err) With(a ...any) error {
	return fmt.Errorf("%w: %s", e, fmt.Sprint(a...))
}

// Withf returns the error with formatted context appended. As with
// fmt.Errorf, the format may use %w to wrap errors, so that they can still be
// matched with errors.Is and errors.As.
func (e Err) Withf(format string, a ...any) error {
	return fmt.Errorf("%w: %w", e, fmt.Errorf(format, a...))
}

// NewDatabaseError returns the driver's error, which is ErrDatabase, and the
// kinds of error it is, such as ErrConflict and ErrUniqueViolation, for a
// Driver to return from its Error method
func NewDatabaseError(err error, kinds ...Err) error {
	if err == nil {
		return nil
	}
	return &DatabaseError{err: err, kinds: append([]Err{ErrDatabase}, kinds...)}
}

// Error returns the driver's error's message
func (e *DatabaseError) Error() string {
	return e.err.Error()
}

// Unwrap returns the driver's error
func (e *DatabaseError) Unwrap() error {
	return e.err
}

// Is returns true when the error is of the kind, or the driver's error is
func (e *DatabaseError) Is(target error) bool {
	if kind, ok := target.(Err); ok && slices.Contains(e.kinds, kind) {
		return true
	}
	return errors.Is(e.err, target)
}

// HTTPError returns the error as an HTTP error, of its status.
func HTTPError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrNotFound):
		return httpresponse.ErrNotFound.With(err.Error())
	case errors.Is(err, ErrBadParameter):
		return httpresponse.ErrBadRequest.With(err.Error())
	case errors.Is(err, ErrNotImplemented):
		return httpresponse.ErrNotImplemented.With(err.Error())
	case errors.Is(err, ErrConflict):
		return httpresponse.ErrConflict.With(err.Error())
	case errors.Is(err, ErrNotAvailable):
		return httpresponse.ErrServiceUnavailable.With(err.Error())
	default:
		return httpresponse.ErrInternalError.With(err.Error())
	}
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// normalize returns the error as one of the package's: no rows is
// ErrNotFound, and the driver's errors are DatabaseErrors
func normalize(driver Driver, err error) error {
	var dbErr *DatabaseError
	var kind Err
	switch {
	case err == nil:
		return nil
	case errors.As(err, &kind), errors.As(err, &dbErr):
		return err
	case errors.Is(err, sql.ErrNoRows):
		return ErrNotFound
	}
	return driver.Error(err)
}
