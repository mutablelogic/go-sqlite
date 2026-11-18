package sqlite

import (
	"errors"
	"testing"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	assert "github.com/stretchr/testify/assert"
)

func Test_HTTPError(t *testing.T) {
	assert := assert.New(t)
	for kind, status := range map[Err]error{
		ErrNotFound:       httpresponse.ErrNotFound,
		ErrBadParameter:   httpresponse.ErrBadRequest,
		ErrNotImplemented: httpresponse.ErrNotImplemented,
		ErrConflict:       httpresponse.ErrConflict,
		ErrNotAvailable:   httpresponse.ErrServiceUnavailable,
		ErrDatabase:       httpresponse.ErrInternalError,
	} {
		assert.ErrorIs(HTTPError(kind.With("thing")), status, kind.Error())
	}
	assert.NoError(HTTPError(nil))
}

func Test_Err_Withf(t *testing.T) {
	assert := assert.New(t)
	cause := errors.New("cause")

	// The error is the code, and wraps an error with %w
	err := ErrBadParameter.Withf("invalid %q: %w", "name", cause)
	assert.ErrorIs(err, ErrBadParameter)
	assert.ErrorIs(err, cause)
	assert.Equal(`bad parameter: invalid "name": cause`, err.Error())

	// Without %w, the message is the same, and nothing else is wrapped
	err = ErrBadParameter.Withf("invalid %q: %v", "name", cause)
	assert.ErrorIs(err, ErrBadParameter)
	assert.NotErrorIs(err, cause)
	assert.Equal(`bad parameter: invalid "name": cause`, err.Error())
}
