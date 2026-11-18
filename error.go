package sqlite

import (
	"errors"
	"fmt"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Err is a kind of error, which SQLite's errors are mapped to
type Err int

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
