package sqlite

import (
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	readers int
	pragmas []Pragma
	bind    *Bind
}

// Opt is an option of a database, which Open is given
type Opt func(*opt) error

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// DefaultReaders is the most connections which read at once
	DefaultReaders = 4

	// DefaultBusyTimeout is how long, in milliseconds, a connection waits
	// for another which is writing
	DefaultBusyTimeout = "10000"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func apply(opts ...Opt) (*opt, error) {
	o := &opt{readers: DefaultReaders, bind: NewBind()}
	for _, fn := range opts {
		if err := fn(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithMaxReaders sets the most connections which read at once, which is
// DefaultReaders, unless it's set. A database in memory has one connection.
func WithMaxReaders(n int) Opt {
	return func(o *opt) error {
		if n < 1 {
			return ErrBadParameter.Withf("readers: %d", n)
		}
		o.readers = n
		return nil
	}
}

// WithPragma sets a pragma on each connection, such as cache_size, after
// the defaults: journal_mode WAL, synchronous NORMAL, foreign_keys ON and
// busy_timeout, which it can change.
func WithPragma(name, value string) Opt {
	return func(o *opt) error {
		if name = strings.TrimSpace(name); name == "" || !isIdentifier(name) {
			return ErrBadParameter.Withf("pragma: %q", name)
		}
		o.pragmas = append(o.pragmas, Pragma{Name: name, Value: value})
		return nil
	}
}

// WithBind sets a bind variable for each query.
func WithBind(k string, v any) Opt {
	return func(o *opt) error {
		if k == "" {
			return ErrBadParameter.With("bind: empty key")
		}
		o.bind.Set(k, v)
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// isIdentifier returns true when the name is letters, digits and
// underscores, which doesn't start with a digit
func isIdentifier(name string) bool {
	for i := 0; i < len(name); i++ {
		c := name[i]
		if !(c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (i > 0 && c >= '0' && c <= '9')) {
			return false
		}
	}
	return name != ""
}
