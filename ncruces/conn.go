package ncruces

import (
	"context"
	"errors"
	"fmt"
	"sync"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	sqlite3 "github.com/ncruces/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// conn is a connection, which writes on the pool's writer and reads on its
// readers, or, in a transaction, does both on the writer, which it holds,
// with its bound parameters
type conn struct {
	pool  *pool
	tx    *tx // the transaction, or nil
	depth int // the savepoints, within the transaction
	bind  *sqlite.Bind
}

// tx is a transaction, on the writer, which one goroutine uses at a time,
// and, of a bulk operation, its prepared statements
type tx struct {
	sync.Mutex
	db    *sqlite3.Conn
	stmts statements
}

// Ensure interfaces are satisfied
var _ sqlite.Conn = (*conn)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a new connection with new bound parameters
func (c *conn) With(params ...any) sqlite.Conn {
	child := *c
	child.bind = c.bind.Copy(params...)
	return &child
}

// Return a new connection with bound queries
func (c *conn) WithQueries(queries ...*sqlite.Queries) sqlite.Conn {
	child := *c
	child.bind = c.bind.WithQueries(queries...)
	return &child
}

// Perform a transaction, which begins immediately, then commit it, or roll it
// back when the function returns an error, or panics. Within a transaction,
// it's a savepoint, which is released, or rolled back to.
func (c *conn) Tx(ctx context.Context, fn func(sqlite.Conn) error) error {
	return c.begin(ctx, fn, false)
}

// Perform a bulk operation within a function, in a transaction, as Tx does,
// whose statements are prepared once, and reused, such as an insert of many
// rows. Within a transaction, it's a savepoint, whose statements are reused
// when the transaction's are.
func (c *conn) Bulk(ctx context.Context, fn func(sqlite.Conn) error) error {
	return c.begin(ctx, fn, true)
}

// begin performs a transaction, or a savepoint, within one, whose statements
// are reused, of a bulk operation
func (c *conn) begin(ctx context.Context, fn func(sqlite.Conn) error, bulk bool) error {
	if c.tx != nil {
		return c.savepoint(ctx, fn)
	}
	db, release, err := c.pool.acquireWriter(ctx)
	if err != nil {
		return err
	}
	defer release()

	// Begin, interrupted when the context is done
	old := db.SetInterrupt(ctx)
	defer db.SetInterrupt(old)
	if err := c.pool.exec(ctx, db, "BEGIN IMMEDIATE"); err != nil {
		return c.pool.error(ctx, err)
	}

	// Roll back, without being interrupted, unless it's committed
	committed := false
	defer func() {
		if !committed {
			db.SetInterrupt(context.Background())
			if !db.GetAutocommit() {
				_ = c.pool.exec(context.WithoutCancel(ctx), db, "ROLLBACK")
			}
		}
	}()
	t := &tx{db: db}
	if bulk {
		t.stmts = statements{}
		defer t.stmts.close()
	}
	if err := fn(&conn{pool: c.pool, tx: t, bind: c.bind.Copy()}); err != nil {
		return c.pool.error(ctx, err)
	}

	// A transaction whose context is done is rolled back, rather than
	// committed, though the function returned nil
	if err := ctx.Err(); err != nil {
		return err
	}
	db.SetInterrupt(ctx)
	if err := c.pool.exec(ctx, db, "COMMIT"); err != nil {
		return c.pool.error(ctx, err)
	}
	committed = true
	return nil
}

// Execute a query, on the writer
func (c *conn) Exec(ctx context.Context, query string) error {
	bind := c.bind.Copy()
	return c.do(ctx, true, func(db *sqlite3.Conn) error {
		return c.execute(ctx, db, bind, query, nil, false)
	})
}

// Perform an insert, binding parameters from the writer, and scanning the
// result into the reader
func (c *conn) Insert(ctx context.Context, reader sqlite.Reader, writer sqlite.Writer) error {
	bind := c.bind.Copy()
	query, err := writer.Insert(bind)
	if err != nil {
		return err
	}
	return c.do(ctx, true, func(db *sqlite3.Conn) error {
		return c.execute(ctx, db, bind, query, reader, false)
	})
}

// Perform an update, selecting using the selector, binding parameters from
// the writer, and scanning the result into the reader
func (c *conn) Update(ctx context.Context, reader sqlite.Reader, sel sqlite.Selector, writer sqlite.Writer) error {
	bind := c.bind.Copy()
	// Perform the update before the select, so the query can be chosen
	// based on the update
	if writer != nil {
		if err := writer.Update(bind); err != nil {
			return err
		}
	}
	query, err := sel.Select(bind, sqlite.Update)
	if err != nil {
		return err
	}
	return c.do(ctx, true, func(db *sqlite3.Conn) error {
		return c.execute(ctx, db, bind, query, reader, false)
	})
}

// Perform a delete, binding parameters with the selector and scanning the
// deleted data into the reader
func (c *conn) Delete(ctx context.Context, reader sqlite.Reader, sel sqlite.Selector) error {
	bind := c.bind.Copy()
	query, err := sel.Select(bind, sqlite.Delete)
	if err != nil {
		return err
	}
	return c.do(ctx, true, func(db *sqlite3.Conn) error {
		return c.execute(ctx, db, bind, query, reader, false)
	})
}

// Perform a get, binding parameters with the selector and scanning its first
// row into the reader, and not the others
func (c *conn) Get(ctx context.Context, reader sqlite.Reader, sel sqlite.Selector) error {
	bind := c.bind.Copy()
	query, err := sel.Select(bind, sqlite.Get)
	if err != nil {
		return err
	}
	return c.do(ctx, false, func(db *sqlite3.Conn) error {
		return c.execute(ctx, db, bind, query, reader, true)
	})
}

// Perform a list, binding parameters with the selector and scanning rows
// into the reader, with ${offsetlimit} after the query, and their count,
// without it, when the reader is a ListReader, of the same snapshot. No
// rows isn't an error.
func (c *conn) List(ctx context.Context, reader sqlite.Reader, sel sqlite.Selector) error {
	bind := c.bind.Copy()
	bind.Set("offsetlimit", "")
	query, err := sel.Select(bind, sqlite.List)
	if err != nil {
		return err
	}

	// It's one statement, without a ; or a comment at its end, which is
	// within the count, and followed by the offset and limit
	query = sqlite.Statement(query)
	return c.do(ctx, false, func(db *sqlite3.Conn) (err error) {
		// The count and the rows are read in one transaction, unless they're
		// in one already
		if c.tx == nil {
			if err := c.pool.exec(ctx, db, "BEGIN"); err != nil {
				return err
			}
			defer func() {
				// It only read, so it ends without being interrupted
				db.SetInterrupt(context.Background())
				if !db.GetAutocommit() {
					err = errors.Join(err, c.pool.exec(context.WithoutCancel(ctx), db, "COMMIT"))
				}
			}()
		}

		// Count the rows, of all of the query
		if counter, ok := reader.(sqlite.ListReader); ok {
			if err := c.execute(ctx, db, bind, `SELECT COUNT(*) FROM (`+query+`)`, readerFunc(counter.ScanCount), false); err != nil {
				return err
			}
		}

		// Read the rows, of the offset and limit
		if err := c.execute(ctx, db, bind, query+` ${offsetlimit}`, reader, false); !errors.Is(err, sqlite.ErrNotFound) {
			return err
		}
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// do calls the function with the transaction's connection, or the writer, or
// a reader, of the pool, which is interrupted when the context is done
func (c *conn) do(ctx context.Context, write bool, fn func(*sqlite3.Conn) error) error {
	var db *sqlite3.Conn
	if c.tx != nil {
		c.tx.Lock()
		defer c.tx.Unlock()
		db = c.tx.db
	} else {
		var release func()
		var err error
		if write {
			db, release, err = c.pool.acquireWriter(ctx)
		} else {
			db, release, err = c.pool.acquireReader(ctx)
		}
		if err != nil {
			return err
		}
		defer release()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	old := db.SetInterrupt(ctx)
	defer db.SetInterrupt(old)
	return c.pool.error(ctx, fn(db))
}

// savepoint performs a function within a savepoint, of the transaction,
// which is released, or rolled back to, and released, when the function
// returns an error, or panics
func (c *conn) savepoint(ctx context.Context, fn func(sqlite.Conn) error) error {
	name := fmt.Sprintf("sp%d", c.depth+1)
	exec := func(ctx context.Context, sql string) error {
		return c.do(ctx, true, func(db *sqlite3.Conn) error { return c.pool.exec(ctx, db, sql) })
	}
	mark := c.pool.changes.savepoint()
	if err := exec(ctx, "SAVEPOINT "+name); err != nil {
		return err
	}

	// Roll back to it, without being interrupted, unless it's released
	released := false
	rollback := func() error {
		ctx := context.WithoutCancel(ctx)
		err := exec(ctx, "ROLLBACK TO "+name)
		c.pool.changes.rollbackTo(mark)
		return errors.Join(err, exec(ctx, "RELEASE "+name))
	}
	defer func() {
		if !released {
			_ = rollback()
		}
	}()
	err := fn(&conn{pool: c.pool, tx: c.tx, depth: c.depth + 1, bind: c.bind.Copy()})
	if err == nil {
		// A savepoint whose context is done is rolled back to, rather than
		// released, though the function returned nil
		err = ctx.Err()
	}
	if err != nil {
		released = true
		return errors.Join(c.pool.error(ctx, err), rollback())
	}
	if err := exec(ctx, "RELEASE "+name); err != nil {
		return err
	}
	released = true
	return nil
}

// execute executes the query, with its ${key} replaced, and its arguments
// bound, which is traced, and scans its rows, or the first, into the reader
func (c *conn) execute(ctx context.Context, db *sqlite3.Conn, bind *sqlite.Bind, query string, reader sqlite.Reader, one bool) error {
	sql, vars := bind.Expand(query)
	name, _ := vars[sqlite.TraceSpanNameArg].(string)
	return c.pool.trace(ctx, name, sql, func(args map[string]any) error {
		return executeRows(db, c.cache(), c.pool.changes, sql, vars, reader, one, args)
	})
}

// cache returns the transaction's prepared statements, of a bulk operation,
// or nil
func (c *conn) cache() statements {
	if c.tx == nil {
		return nil
	}
	return c.tx.stmts
}

// readerFunc is a function which scans a row
type readerFunc func(sqlite.Row) error

func (fn readerFunc) Scan(row sqlite.Row) error {
	return fn(row)
}
