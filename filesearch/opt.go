package filesearch

import (
	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is an option of a manager, when it's created
type Opt func(*opt) error

// opt is the options of a manager
type opt struct {
	tracer  trace.Tracer
	readers int
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (o *opt) apply(opts []Opt) error {
	o.readers = sqlite.DefaultReaders
	for _, fn := range opts {
		if err := fn(o); err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithTracer sets the OpenTelemetry tracer, of the spans of the manager,
// and of its database's queries
func WithTracer(tracer trace.Tracer) Opt {
	return func(o *opt) error {
		o.tracer = tracer
		return nil
	}
}

// WithMaxReaders sets the most connections which read the database at
// once, such as to search it, while it's indexed
func WithMaxReaders(n int) Opt {
	return func(o *opt) error {
		if n < 1 {
			return sqlite.ErrBadParameter.Withf("readers: %d", n)
		}
		o.readers = n
		return nil
	}
}
