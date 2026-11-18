// Package ncruces opens SQLite databases with ncruces/go-sqlite3, which is
// SQLite compiled to WebAssembly, so it's pure Go, without cgo.
package ncruces

import (
	"context"
	"errors"
	"strings"
	"sync"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	sqlite3 "github.com/ncruces/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// pool is a database, which writes on one connection, as SQLite writes one
// transaction at a time, and reads on many, which see what's committed, as
// a file's journal is a write-ahead log
type pool struct {
	opts    *sqlite.Options
	init    []func(*sqlite3.Conn) error
	writer  *sqlite3.Conn
	write   chan struct{}      // the writer's token, which is held while it's used
	readers chan *sqlite3.Conn // the readers, which are idle, or nil in memory
	count   int                // the readers
	done    chan struct{}      // closed when the pool is
	close   sync.Once
}

// poolconn is a pool's connection, with its bound parameters
type poolconn struct {
	conn
}

// initOpt is the functions which are called on each connection
type initOpt []func(*sqlite3.Conn) error

// Ensure interfaces are satisfied
var _ sqlite.PoolConn = (*poolconn)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open opens the database of the path, which is created when it doesn't
// exist, or of sqlite.Memory.
//
// It's written on one connection, so a transaction, or a write, waits for
// another to end, and read on many, of sqlite.WithMaxReaders, which don't
// wait. A file's journal is a write-ahead log, so they read what's committed
// while it's written. Within a transaction, use its Conn, rather than the
// pool's, to write, as the pool's waits for the transaction to end.
func Open(ctx context.Context, path string, opts ...sqlite.Opt) (sqlite.PoolConn, error) {
	o, err := sqlite.Apply(opts...)
	if err != nil {
		return nil, err
	}
	if path = strings.TrimSpace(path); path == "" {
		return nil, sqlite.ErrBadParameter.With("path is empty")
	}
	p := &pool{opts: o, write: make(chan struct{}, 1), done: make(chan struct{})}
	for _, opt := range o.Driver {
		fn, ok := opt.(initOpt)
		if !ok {
			return nil, sqlite.ErrBadParameter.Withf("option of another implementation: %T", opt)
		}
		p.init = append(p.init, fn...)
	}
	memory := path == sqlite.Memory

	// The writer, whose file is created, and its journal set, before it's
	// read
	pragmas := []sqlite.Pragma{{Name: "foreign_keys", Value: "ON"}, {Name: "busy_timeout", Value: sqlite.DefaultBusyTimeout}}
	if !memory {
		pragmas = append(pragmas, sqlite.Pragma{Name: "journal_mode", Value: "WAL"}, sqlite.Pragma{Name: "synchronous", Value: "NORMAL"})
	}
	if p.writer, err = p.open(ctx, path, false, append(pragmas, o.Pragmas...)); err != nil {
		return nil, err
	}
	p.write <- struct{}{}

	// The readers, which only query, of a file
	if !memory {
		p.readers = make(chan *sqlite3.Conn, o.Readers)
		pragmas := []sqlite.Pragma{{Name: "busy_timeout", Value: sqlite.DefaultBusyTimeout}, {Name: "query_only", Value: "ON"}}
		for range o.Readers {
			reader, err := p.open(ctx, path, true, append(pragmas, o.Pragmas...))
			if err != nil {
				return nil, errors.Join(err, p.Close())
			}
			p.readers <- reader
			p.count++
		}
	}

	// Return the pool's connection
	return &poolconn{conn{pool: p, bind: o.Bind}}, nil
}

////////////////////////////////////////////////////////////////////////////////
// OPTIONS

// WithInit calls the functions, in order, on each connection as it's opened,
// such as to create collations or virtual tables, or to register an
// extension, of github.com/ncruces/go-sqlite3/ext, whose Register functions
// are of the same signature. Functions are created with sqlite.WithFunction.
func WithInit(fn ...func(*sqlite3.Conn) error) sqlite.Opt {
	var init initOpt
	for _, fn := range fn {
		if fn != nil {
			init = append(init, fn)
		}
	}
	return sqlite.WithDriverOption(init)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Ping the writer, and a reader
func (p *poolconn) Ping(ctx context.Context) error {
	if err := p.do(ctx, true, ping); err != nil {
		return err
	}
	return p.do(ctx, false, ping)
}

// Close the database, once its connections are idle
func (p *poolconn) Close() error {
	return p.pool.Close()
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// open opens a connection, which reads and writes, creating its file, or only
// reads, with its pragmas, functions and init functions, which are
// interrupted when the context is done
func (p *pool) open(ctx context.Context, path string, readonly bool, pragmas []sqlite.Pragma) (*sqlite3.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	flags := sqlite3.OPEN_READWRITE | sqlite3.OPEN_CREATE
	if readonly {
		flags = sqlite3.OPEN_READONLY
	}
	db, err := sqlite3.OpenFlags(path, flags)
	if err != nil {
		return nil, p.error(ctx, err)
	}
	db.SetInterrupt(ctx)
	if err := p.setup(db, pragmas); err != nil {
		return nil, errors.Join(p.error(ctx, err), db.Close())
	}
	db.SetInterrupt(context.Background())
	return db, nil
}

// setup sets a connection's pragmas, and creates its functions
func (p *pool) setup(db *sqlite3.Conn, pragmas []sqlite.Pragma) error {
	for _, pragma := range pragmas {
		if err := db.Exec("PRAGMA " + pragma.Name + " = " + pragma.Value); err != nil {
			return err
		}
	}
	for _, fn := range p.opts.Functions {
		if err := createFunction(db, fn); err != nil {
			return err
		}
	}
	for _, fn := range p.init {
		if err := fn(db); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the readers and the writer, once they're idle, and then
// they can't be used
func (p *pool) Close() error {
	var result error
	p.close.Do(func() {
		close(p.done)
		for range p.count {
			result = errors.Join(result, (<-p.readers).Close())
		}
		if p.writer != nil {
			<-p.write
			result = errors.Join(result, p.writer.Close())
		}
	})
	return result
}

// writer waits for the writer, until the context is done, or the pool is
// closed, and returns it, and the function which returns it
func (p *pool) acquireWriter(ctx context.Context) (*sqlite3.Conn, func(), error) {
	if err := p.ready(ctx); err != nil {
		return nil, nil, err
	}
	select {
	case <-p.write:
		return p.writer, func() { p.write <- struct{}{} }, nil
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case <-p.done:
		return nil, nil, sqlite.ErrNotAvailable.With("database is closed")
	}
}

// acquireReader waits for a reader, or the writer, in memory, until the
// context is done, or the pool is closed, and returns it, and the function
// which returns it
func (p *pool) acquireReader(ctx context.Context) (*sqlite3.Conn, func(), error) {
	if p.readers == nil {
		return p.acquireWriter(ctx)
	}
	if err := p.ready(ctx); err != nil {
		return nil, nil, err
	}
	select {
	case reader := <-p.readers:
		return reader, func() { p.readers <- reader }, nil
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case <-p.done:
		return nil, nil, sqlite.ErrNotAvailable.With("database is closed")
	}
}

// ready returns an error when the context is done, or the pool is closed
func (p *pool) ready(ctx context.Context) error {
	select {
	case <-p.done:
		return sqlite.ErrNotAvailable.With("database is closed")
	default:
		return ctx.Err()
	}
}

// ping queries the connection
func ping(db *sqlite3.Conn) error {
	return db.Exec("SELECT 1")
}
