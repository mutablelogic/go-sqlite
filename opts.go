package sqlite

import (
	"regexp"
	"strings"

	// Packages
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	values "github.com/mutablelogic/go-sqlite/values"
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

	// The broadcaster, which the rows which are written are published to,
	// once they're committed, and the tables which are, or all of them
	Broadcaster Publisher
	Tables      []string

	// The schema of the changes of the main database, which are published,
	// rather than main
	Schema string
}

// Publisher is what changes are published to, such as go-server's
// broadcaster.Broadcaster. Publish doesn't block, and returns the watchers
// which missed the change.
type Publisher interface {
	Publish(schema.Change) int
}

// Opt is an option of a database, which Open is given
type Opt func(*Options) error

// Pragma is a pragma, which is set on each connection, such as
// cache_size, of -4000.
type Pragma struct {
	Name  string
	Value string
}

// Function is a function, which SQL calls by its name, which an
// implementation creates on each connection
type Function struct {
	Name          string
	Args          int  // The number of arguments, or -1 for any
	Deterministic bool // The same arguments always return the same result

	// Fn is called with the arguments, which it scans, and returns its
	// result, which is set as values.Set sets it, or an error
	Fn func(args []values.Value) (any, error)
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// MaxChanges is the most rows of a table, of a transaction, which are
	// published as changes of each, rather than one change of the table
	MaxChanges = 1000

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
// its name, of Func0, Func1, Func2, Func3 or FuncN, such as:
//
//	sqlite.WithFunction("double", true, sqlite.Func1(func(n int64) (int64, error) {
//		return n * 2, nil
//	}))
//
// A deterministic function always returns the same result for the same
// arguments, so it can be used in indexes.
func WithFunction(name string, deterministic bool, fn Fn) Opt {
	return func(o *Options) error {
		if !isIdentifier(name) {
			return ErrBadParameter.Withf("function: %q", name)
		}
		if fn.fn == nil {
			return ErrBadParameter.Withf("function %s: nil", name)
		}
		o.Functions = append(o.Functions, Function{Name: name, Args: fn.args, Deterministic: deterministic, Fn: fn.fn})
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

// WithBroadcaster publishes the rows which are inserted, updated and deleted,
// once they're committed, to a broadcaster, such as go-server's
// broadcaster.Broadcaster, as a change of each row: its action, the
// database, such as main, its table, and its rowid. Changes of the tables,
// when there are some, are published, or else of all of them, other than
// the shadow tables of virtual tables, such as fts5's. A row which is
// written more than once in a transaction is one change. A transaction of
// more than MaxChanges rows of a table is one change of the table, without
// a row. Writes by other processes, and of tables WITHOUT ROWID, aren't
// published.
//
// Publish is called after the commit, before the writer is released, so it
// mustn't block, or query the database.
func WithBroadcaster(b Publisher, tables ...string) Opt {
	return func(o *Options) error {
		if b == nil {
			return ErrBadParameter.With("broadcaster: nil")
		}
		for _, table := range tables {
			if !isIdentifier(table) {
				return ErrBadParameter.Withf("broadcaster: table %q", table)
			}
		}
		o.Broadcaster, o.Tables = b, append(o.Tables, tables...)
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

// WithSchemaName sets the schema of the changes of the main database,
// which are published to the broadcaster, with WithBroadcaster, rather than
// main, such as webhook, so the changes of databases whose tables have the
// same names, which are published to one broadcaster, are told apart. The
// changes of attached databases are of their names.
func WithSchemaName(name string) Opt {
	return func(o *Options) error {
		if !isIdentifier(name) {
			return ErrBadParameter.Withf("schema name: %q", name)
		}
		o.Schema = name
		return nil
	}
}

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
