package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// pool is a database, which writes on one connection, as SQLite writes one
// transaction at a time, and reads on many, which see what's committed, as
// its journal is a write-ahead log
type pool struct {
	driver Driver
	writer *sql.DB
	reader *sql.DB // the writer, of a database in memory
}

// poolconn is a pool's connection, with its bound parameters
type poolconn struct {
	conn
}

// Ensure interfaces are satisfied
var _ PoolConn = (*poolconn)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Memory is the path of a database in memory, which is one connection's,
	// so it's both written and read on one
	Memory = ":memory:"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open opens the database of the path, which is created when it doesn't
// exist, or of Memory, with the driver, such as ncruces's.
//
// It's written on one connection, so a transaction, or a write, waits for
// another to end, and read on many, of WithMaxReaders, which don't wait. A
// file's journal is a write-ahead log, so they read what's committed while
// it's written. Within a transaction, use its Conn, rather than the pool's,
// to write, as the pool's waits for the transaction to end.
func Open(ctx context.Context, path string, driver Driver, opts ...Opt) (PoolConn, error) {
	if driver == nil {
		return nil, ErrBadParameter.With("driver is nil")
	}
	if path = strings.TrimSpace(path); path == "" {
		return nil, ErrBadParameter.With("path is empty")
	}
	o, err := apply(opts...)
	if err != nil {
		return nil, err
	}
	memory := path == Memory

	// The writer, of one connection, which is kept open, as a database in
	// memory is closed with it
	pragmas := []Pragma{{"foreign_keys", "ON"}, {"busy_timeout", DefaultBusyTimeout}}
	if !memory {
		pragmas = append(pragmas, Pragma{"journal_mode", "WAL"}, Pragma{"synchronous", "NORMAL"})
	}
	writer, err := driver.Open(path, false, append(pragmas, o.pragmas...))
	if err != nil {
		return nil, normalize(driver, err)
	}
	writer.SetMaxOpenConns(1)
	writer.SetMaxIdleConns(1)
	writer.SetConnMaxLifetime(0)
	writer.SetConnMaxIdleTime(0)
	p := &pool{driver: driver, writer: writer, reader: writer}

	// The file is created, and its journal set, before it's read
	if err := writer.PingContext(ctx); err != nil {
		return nil, errors.Join(normalize(driver, err), writer.Close())
	}

	// The readers, which only query, of a file
	if !memory {
		pragmas := []Pragma{{"busy_timeout", DefaultBusyTimeout}, {"query_only", "ON"}}
		reader, err := driver.Open(path, true, append(pragmas, o.pragmas...))
		if err != nil {
			return nil, errors.Join(normalize(driver, err), writer.Close())
		}
		reader.SetMaxOpenConns(o.readers)
		reader.SetMaxIdleConns(o.readers)
		p.reader = reader
		if err := reader.PingContext(ctx); err != nil {
			return nil, errors.Join(normalize(driver, err), p.close())
		}
	}

	// Return the pool's connection
	return &poolconn{conn{pool: p, write: writer, read: p.reader, bind: o.bind}}, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Ping the writer and the readers
func (p *poolconn) Ping(ctx context.Context) error {
	if err := p.pool.writer.PingContext(ctx); err != nil {
		return normalize(p.pool.driver, err)
	}
	if p.pool.reader != p.pool.writer {
		return normalize(p.pool.driver, p.pool.reader.PingContext(ctx))
	}
	return nil
}

// Close the readers, and then the writer
func (p *poolconn) Close() error {
	return p.pool.close()
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *pool) close() error {
	var result error
	if p.reader != p.writer {
		result = errors.Join(result, p.reader.Close())
	}
	return errors.Join(result, p.writer.Close())
}
