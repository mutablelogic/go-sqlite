package sqlite

import (
	"regexp"
	"strings"

	// Packages
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Options are what Open is given, which an implementation, such as package
// ncruces, reads, with Apply
type Options struct {
	// The most connections which read at once
	Readers int

	// The pragmas, which are set on each connection, after the defaults
	Pragmas []Pragma

	// The functions, which are created on each connection
	Functions []Function

	// The bind variables of each query
	Bind *Bind

	// The OpenTelemetry tracer, of a span of each query
	Tracer trace.Tracer
}

// Opt is an option of a database, which Open is given
type Opt func(*Options) error

// Pragma is a pragma, which is set on each connection, such as
// cache_size, of -4000.
type Pragma struct {
	Name  string
	Value string
}

// Function is a function, which SQL calls, with its arguments, which are
// int64, float64, string, []byte or nil, and returns its result, which is a
// bool, an integer, a float, a string, a []byte, a time.Time, a
// driver.Valuer or nil, or an error.
type Function struct {
	Name          string
	Args          int  // The number of arguments, or -1 for any
	Deterministic bool // The same arguments always return the same result
	Fn            func(args ...any) (any, error)
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// DefaultReaders is the most connections which read at once
	DefaultReaders = 4

	// DefaultBusyTimeout is how long, in milliseconds, a connection waits
	// for another which is writing
	DefaultBusyTimeout = "10000"

	// Memory is the path of a database in memory, which is one connection's,
	// so it's both written and read on one
	Memory = ":memory:"
)

var (
	// A pragma's value is a number, a name, or a string
	rePragmaValue = regexp.MustCompile(`^(-?[0-9]+|[A-Za-z_][A-Za-z0-9_]*|'[^']*')$`)
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Apply returns the options, of their defaults, as they're set
func Apply(opts ...Opt) (*Options, error) {
	o := &Options{Readers: DefaultReaders, Bind: NewBind()}
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
	return func(o *Options) error {
		if n < 1 {
			return ErrBadParameter.Withf("readers: %d", n)
		}
		o.Readers = n
		return nil
	}
}

// WithPragma sets a pragma on each connection, such as cache_size, after
// the defaults: journal_mode WAL, synchronous NORMAL, foreign_keys ON and
// busy_timeout, which it can change. Its value is a number, a name, or a
// string in single quotes.
func WithPragma(name, value string) Opt {
	name, value = strings.TrimSpace(name), strings.TrimSpace(value)
	return func(o *Options) error {
		if !isIdentifier(name) {
			return ErrBadParameter.Withf("pragma: %q", name)
		}
		if !rePragmaValue.MatchString(value) {
			return ErrBadParameter.Withf("pragma %s: %q", name, value)
		}
		o.Pragmas = append(o.Pragmas, Pragma{Name: name, Value: value})
		return nil
	}
}

// WithFunction creates a function on each connection, which SQL calls by
// its name, with the number of arguments, or -1 for any. A deterministic
// function always returns the same result for the same arguments, so it can
// be used in indexes.
func WithFunction(name string, args int, deterministic bool, fn func(args ...any) (any, error)) Opt {
	return func(o *Options) error {
		if !isIdentifier(name) {
			return ErrBadParameter.Withf("function: %q", name)
		}
		if args < -1 || args > 127 {
			return ErrBadParameter.Withf("function %s: %d arguments", name, args)
		}
		if fn == nil {
			return ErrBadParameter.Withf("function %s: nil", name)
		}
		o.Functions = append(o.Functions, Function{Name: name, Args: args, Deterministic: deterministic, Fn: fn})
		return nil
	}
}

// WithTracer sets the OpenTelemetry tracer, of a span of each query, which
// is named by the query's name, of Bind.Query, or sqlite.query, or, of a
// transaction's statements, sqlite.transaction, sqlite.commit or
// sqlite.rollback, with its SQL, and its arguments, as arg. and their
// names, and of opening the database, sqlite.open.
func WithTracer(t trace.Tracer) Opt {
	return func(o *Options) error {
		o.Tracer = t
		return nil
	}
}

// WithBind sets a bind variable for each query.
func WithBind(k string, v any) Opt {
	return func(o *Options) error {
		if k == "" {
			return ErrBadParameter.With("bind: empty key")
		}
		o.Bind.Set(k, v)
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
