// Package filesearch indexes the files of volumes, which are folders, of
// their metadata, content and artwork, and searches them, in a SQLite
// database, which it creates, and which is swept of what's gone, when a
// volume is indexed again.
package filesearch

import (
	"context"
	"errors"
	"sync"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	types "github.com/mutablelogic/go-server/pkg/types"
	sqlite "github.com/mutablelogic/go-sqlite"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
	ncruces "github.com/mutablelogic/go-sqlite/ncruces"
	attribute "go.opentelemetry.io/otel/attribute"

	// The schema searches objects' text, with fts5
	_ "github.com/mutablelogic/go-sqlite/ncruces/fts5"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Manager indexes and searches the files of volumes, in its database
type Manager struct {
	opt
	pool sqlite.PoolConn
	conn sqlite.Conn // The pool's connection, with the schema's queries

	// The progress of the volumes which are indexing, or were, by their paths
	sync.Mutex
	progress map[string]*progress
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New returns a manager, of the database of the path, which is created when
// it doesn't exist, or sqlite.Memory, and creates its schema, when it
// hasn't it.
func New(ctx context.Context, path string, opts ...Opt) (_ *Manager, err error) {
	self := &Manager{progress: make(map[string]*progress)}
	if err := self.opt.apply(opts); err != nil {
		return nil, err
	}

	// Trace creating the manager, which needs the tracer from the options
	ctx, endSpan := otel.StartSpan(self.tracer, ctx, "filesearch.New",
		attribute.String("arg.path", path),
	)
	defer func() { endSpan(err) }()

	// Open the database
	pool, err := ncruces.Open(ctx, path, sqlite.WithTracer(self.tracer), sqlite.WithMaxReaders(self.readers))
	if err != nil {
		return nil, err
	}

	// Bootstrap it, with the schema, and its queries
	queries, err := schema.NewQueries()
	if err != nil {
		return nil, errors.Join(sqlite.ErrInternalServerError.Withf("queries.sql: %w", err), pool.Close())
	}
	if err := schema.Create(ctx, pool); err != nil {
		return nil, errors.Join(err, pool.Close())
	}
	self.pool, self.conn = pool, pool.WithQueries(queries)

	// Return success
	return self, nil
}

// Close closes the database, once its connections are idle
func (m *Manager) Close() error {
	return m.pool.Close()
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Ping the database
func (m *Manager) Ping(ctx context.Context) error {
	return m.pool.Ping(ctx)
}

// ListObjects lists the objects, of a volume, or all of them, by their
// paths, or searches them, when there's a query, the best first
func (m *Manager) ListObjects(ctx context.Context, req schema.ObjectListRequest) (_ *schema.ObjectList, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "filesearch.ListObjects",
		attribute.String("arg.req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()
	var list schema.ObjectList
	if err := m.conn.List(ctx, &list, req); err != nil {
		return nil, err
	}
	return &list, nil
}
