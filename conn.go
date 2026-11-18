// Package sqlite is a SQLite client, of the same shape as go-pg's: queries
// are bound to a connection, and objects are read and written by the
// Reader, Writer and Selector interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// PoolConn is a database, opened from a file or in memory, which writes on
// one connection, and reads on many.
type PoolConn interface {
	Conn

	// Ping the database
	Ping(context.Context) error

	// Close the database
	Close() error
}

// Conn is a connection, with bound parameters and queries.
type Conn interface {
	// Return a new connection with bound parameters
	With(...any) Conn

	// Return a new connection with bound queries
	WithQueries(...*Queries) Conn

	// Perform a transaction within a function, on the connection which
	// writes, which is begun immediately, so it isn't upgraded to write
	// later, when another transaction may be writing
	Tx(context.Context, func(Conn) error) error

	// Perform a bulk operation within a function, in a transaction, whose
	// statements are prepared once
	Bulk(context.Context, func(Conn) error) error

	// Execute a query
	Exec(context.Context, string) error

	// Perform an insert
	Insert(context.Context, Reader, Writer) error

	// Perform an update
	Update(context.Context, Reader, Selector, Writer) error

	// Perform a delete
	Delete(context.Context, Reader, Selector) error

	// Perform a get, on a connection which reads, unless in a transaction
	Get(context.Context, Reader, Selector) error

	// Perform a list, on a connection which reads, unless in a transaction.
	// If the reader is a ListReader, then the count of items is also
	// calculated
	List(context.Context, Reader, Selector) error
}

// querier is what's queried: a database, or a transaction
type querier interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// conn is a connection, which writes on one querier and reads on another,
// or, in a transaction, does both on it, with its bound parameters
type conn struct {
	pool  *pool
	write querier
	read  querier
	tx    *sql.Tx // the transaction, or nil
	depth int     // the savepoints, within the transaction
	bind  *Bind
}

// Ensure interfaces are satisfied
var _ Conn = (*conn)(nil)

// Op represents a database operation type.
type Op uint

// Row is a row of a query's results, such as a *sql.Row or *sql.Rows.
type Row interface {
	Scan(...any) error
}

// Reader scans a database row into an object.
type Reader interface {
	// Scan row into a result
	Scan(Row) error
}

// ListReader scans database rows and counts total results.
type ListReader interface {
	Reader

	// Scan count into the result
	ScanCount(Row) error
}

// Writer binds object fields for insert or update operations.
type Writer interface {
	// Set bind parameters for an insert
	Insert(*Bind) (string, error)

	// Set bind parameters for an update
	Update(*Bind) error
}

// Selector binds parameters for get, update, or delete operations.
type Selector interface {
	// Set bind parameters for getting, updating or deleting
	Select(*Bind, Op) (string, error)
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

// Operations
const (
	None Op = iota
	Get
	Insert
	Update
	Delete
	List
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o Op) String() string {
	switch o {
	case Get:
		return "GET"
	case Insert:
		return "INSERT"
	case Update:
		return "UPDATE"
	case Delete:
		return "DELETE"
	case List:
		return "LIST"
	}
	return "UNKNOWN"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - CONN

// Return a new connection with new bound parameters
func (c *conn) With(params ...any) Conn {
	child := *c
	child.bind = c.bind.Copy(params...)
	return &child
}

// Return a new connection with bound queries
func (c *conn) WithQueries(queries ...*Queries) Conn {
	child := *c
	child.bind = c.bind.withQueries(queries...)
	return &child
}

// Perform a transaction, then commit it, or roll it back when the function
// returns an error, or panics. Within a transaction, it's a savepoint, which
// is released, or rolled back to.
func (c *conn) Tx(ctx context.Context, fn func(Conn) error) (result error) {
	if c.tx != nil {
		return c.savepoint(ctx, fn)
	}
	// The writer is waited for, until the context is done, and then the
	// transaction is begun without it, so it's this which rolls it back, and
	// not database/sql, which closes the connection when it does, so a
	// database in memory, which is the connection's, would be lost. Its
	// statements are interrupted when the context is done.
	writer, err := c.pool.writer.Conn(ctx)
	if err != nil {
		return c.error(err)
	}
	defer writer.Close()
	tx, err := writer.BeginTx(context.WithoutCancel(ctx), nil)
	if err != nil {
		return c.error(err)
	}
	done := false
	defer func() {
		if !done {
			_ = tx.Rollback()
		}
	}()
	err = fn(&conn{pool: c.pool, write: tx, read: tx, tx: tx, bind: c.bind.Copy()})
	done = true
	if err != nil {
		return errors.Join(c.error(err), c.error(ignoreDone(tx.Rollback())))
	}
	return c.error(tx.Commit())
}

// Perform a bulk operation within a function, in a transaction
func (c *conn) Bulk(ctx context.Context, fn func(Conn) error) error {
	return c.Tx(ctx, fn)
}

// Execute a query, on the writer
func (c *conn) Exec(ctx context.Context, query string) error {
	return c.exec(ctx, c.write, c.bind.Copy(), query, nil)
}

// Perform an insert, binding parameters from the writer, and scanning the
// result into the reader
func (c *conn) Insert(ctx context.Context, reader Reader, writer Writer) error {
	bind := c.bind.Copy()
	query, err := writer.Insert(bind)
	if err != nil {
		return err
	}
	return c.exec(ctx, c.write, bind, query, reader)
}

// Perform an update, selecting using the selector, binding parameters from
// the writer, and scanning the result into the reader
func (c *conn) Update(ctx context.Context, reader Reader, sel Selector, writer Writer) error {
	bind := c.bind.Copy()
	// Perform the update before the select, so the query can be chosen
	// based on the update
	if writer != nil {
		if err := writer.Update(bind); err != nil {
			return err
		}
	}
	query, err := sel.Select(bind, Update)
	if err != nil {
		return err
	}
	return c.exec(ctx, c.write, bind, query, reader)
}

// Perform a delete, binding parameters with the selector and scanning the
// deleted data into the reader
func (c *conn) Delete(ctx context.Context, reader Reader, sel Selector) error {
	bind := c.bind.Copy()
	query, err := sel.Select(bind, Delete)
	if err != nil {
		return err
	}
	return c.exec(ctx, c.write, bind, query, reader)
}

// Perform a get, binding parameters with the selector and scanning a single
// row into the reader
func (c *conn) Get(ctx context.Context, reader Reader, sel Selector) error {
	bind := c.bind.Copy()
	query, err := sel.Select(bind, Get)
	if err != nil {
		return err
	}
	return c.exec(ctx, c.read, bind, query, reader)
}

// Perform a list, binding parameters with the selector and scanning rows
// into the reader, with ${offsetlimit} after the query, and their count,
// without it, when the reader is a ListReader. No rows isn't an error.
func (c *conn) List(ctx context.Context, reader Reader, sel Selector) error {
	bind := c.bind.Copy()
	bind.Set("offsetlimit", "")
	query, err := sel.Select(bind, List)
	if err != nil {
		return err
	}

	// Count the rows, of all of the query
	if counter, ok := reader.(ListReader); ok {
		vars := bind.snapshot()
		count := replace(`SELECT COUNT(*) FROM (`+query+`)`, vars)
		if err := counter.ScanCount(c.read.QueryRowContext(ctx, count, args(count, vars)...)); err != nil {
			return c.error(err)
		}
	}

	// Read the rows, of the offset and limit
	if err := c.exec(ctx, c.read, bind, query+` ${offsetlimit}`, reader); !errors.Is(err, ErrNotFound) {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// exec executes the query, with its ${key} replaced, and its @key arguments,
// and scans its rows into the reader, which is ErrNotFound without any
func (c *conn) exec(ctx context.Context, q querier, bind *Bind, query string, reader Reader) error {
	vars := bind.snapshot()
	query = replace(query, vars)

	// Without a reader, just execute the query
	if reader == nil {
		_, err := q.ExecContext(ctx, query, args(query, vars)...)
		return c.error(err)
	}

	// Execute the query, and read its rows
	rows, err := q.QueryContext(ctx, query, args(query, vars)...)
	if err != nil {
		return c.error(err)
	}
	defer rows.Close()
	var scanned bool
	for rows.Next() {
		if err := reader.Scan(rows); err != nil {
			return c.error(err)
		}
		scanned = true
	}
	if err := rows.Err(); err != nil {
		return c.error(err)
	} else if !scanned {
		return ErrNotFound
	}
	return nil
}

// savepoint performs a function within a savepoint, of the transaction,
// which is released, or rolled back to, and released, when the function
// returns an error, or panics
func (c *conn) savepoint(ctx context.Context, fn func(Conn) error) error {
	name := fmt.Sprintf("sp%d", c.depth+1)
	if _, err := c.tx.ExecContext(ctx, "SAVEPOINT "+name); err != nil {
		return c.error(err)
	}
	rollback := func() error {
		ctx := context.WithoutCancel(ctx)
		if _, err := c.tx.ExecContext(ctx, "ROLLBACK TO "+name); err != nil {
			return c.error(err)
		}
		_, err := c.tx.ExecContext(ctx, "RELEASE "+name)
		return c.error(err)
	}
	done := false
	defer func() {
		if !done {
			_ = rollback()
		}
	}()
	err := fn(&conn{pool: c.pool, write: c.tx, read: c.tx, tx: c.tx, depth: c.depth + 1, bind: c.bind.Copy()})
	done = true
	if err != nil {
		return errors.Join(c.error(err), rollback())
	}
	_, err = c.tx.ExecContext(ctx, "RELEASE "+name)
	return c.error(err)
}

// error returns the error as one of the package's
func (c *conn) error(err error) error {
	return normalize(c.pool.driver, err)
}

// ignoreDone returns nil when the transaction was already ended, such as by
// its context being cancelled
func ignoreDone(err error) error {
	if errors.Is(err, sql.ErrTxDone) {
		return nil
	}
	return err
}

// args returns the named arguments of the query, of its @key parameters, as
// they're bound, or NULL when they aren't
func args(query string, vars map[string]any) []any {
	keys := queryArgs(query)
	if len(keys) == 0 {
		return nil
	}
	result := make([]any, 0, len(keys))
	for _, key := range keys {
		result = append(result, sql.Named(key, vars[key]))
	}
	return result
}

// queryArgs returns the names of the query's @key parameters, once each, in
// the order they're first used
func queryArgs(query string) []string {
	var keys []string
	seen := make(map[string]bool)
	for i := 0; i < len(query); i++ {
		if query[i] != '@' || i+1 >= len(query) || !isArgStart(query[i+1]) {
			continue
		}
		j := i + 2
		for j < len(query) && isArgPart(query[j]) {
			j++
		}
		if key := query[i+1 : j]; !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
		i = j - 1
	}
	return keys
}

func isArgStart(ch byte) bool {
	return ch == '_' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}

func isArgPart(ch byte) bool {
	return isArgStart(ch) || ch >= '0' && ch <= '9'
}
