package sqlite

import "fmt"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Error uint

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ErrSuccess Error = iota
	ErrBadParameter
	ErrDuplicateEntry
	ErrUnexpectedResponse
	ErrNotFound
	ErrNotModified
	ErrInternalAppError
	ErrNotImplemented
)

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e Error) Error() string {
	switch e {
	case ErrSuccess:
		return "ErrSuccess"
	case ErrBadParameter:
		return "ErrBadParameter"
	case ErrDuplicateEntry:
		return "ErrDuplicateEntry"
	case ErrUnexpectedResponse:
		return "ErrUnexpectedResponse"
	case ErrNotFound:
		return "ErrNotFound"
	case ErrNotModified:
		return "ErrNotModified"
	case ErrInternalAppError:
		return "ErrInternalAppError"
	case ErrNotImplemented:
		return "ErrNotImplemented"
	default:
		return "[?? Invalid Error value]"
	}
}

func (e Error) With(args ...interface{}) error {
	return fmt.Errorf("%s: %w", fmt.Sprint(args...), e)
}
