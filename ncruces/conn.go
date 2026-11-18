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

// tx is a transaction, on the writer, which one goroutine uses at a time
type tx struct {
	sync.Mutex
	db *sqlite3.Conn
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
	txn, err := db.BeginImmediate()
	if err != nil {
		return c.pool.error(ctx, err)
	}

	// Roll back, without being interrupted, unless it's committed
	committed := false
	defer func() {
		if !committed {
			db.SetInterrupt(context.Background())
			_ = txn.Rollback()
		}
	}()
	if err := fn(&conn{pool: c.pool, tx: &tx{db: db}, bind: c.bind.Copy()}); err != nil {
		return c.pool.error(ctx, err)
	}
	db.SetInterrupt(ctx)
	if err := txn.Commit(); err != nil {
		return c.pool.error(ctx, err)
	}
	committed = true
	return nil
}

// Perform a bulk operation within a function, in a transaction
func (c *conn) Bulk(ctx context.Context, fn func(sqlite.Conn) error) error {
	return c.Tx(ctx, fn)
}

// Execute a query, on the writer
func (c *conn) Exec(ctx context.Context, query string) error {
	bind := c.bind.Copy()
	return c.do(ctx, true, func(db *sqlite3.Conn) error {
		return execute(db, bind, query, nil)
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
		return execute(db, bind, query, reader)
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
		return execute(db, bind, query, reader)
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
		return execute(db, bind, query, reader)
	})
}

// Perform a get, binding parameters with the selector and scanning a single
// row into the reader
func (c *conn) Get(ctx context.Context, reader sqlite.Reader, sel sqlite.Selector) error {
	bind := c.bind.Copy()
	query, err := sel.Select(bind, sqlite.Get)
	if err != nil {
		return err
	}
	return c.do(ctx, false, func(db *sqlite3.Conn) error {
		return execute(db, bind, query, reader)
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
	return c.do(ctx, false, func(db *sqlite3.Conn) (err error) {
		// The count and the rows are read in one transaction, unless they're
		// in one already
		if c.tx == nil {
			if err := db.Exec("BEGIN"); err != nil {
				return err
			}
			defer func() {
				// It only read, so it ends without being interrupted
				db.SetInterrupt(context.Background())
				if !db.GetAutocommit() {
					err = errors.Join(err, db.Exec("COMMIT"))
				}
			}()
		}

		// Count the rows, of all of the query
		if counter, ok := reader.(sqlite.ListReader); ok {
			if err := execute(db, bind, `SELECT COUNT(*) FROM (`+query+`)`, readerFunc(counter.ScanCount)); err != nil {
				return err
			}
		}

		// Read the rows, of the offset and limit
		if err := execute(db, bind, query+` ${offsetlimit}`, reader); !errors.Is(err, sqlite.ErrNotFound) {
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
		return c.do(ctx, true, func(db *sqlite3.Conn) error { return db.Exec(sql) })
	}
	if err := exec(ctx, "SAVEPOINT "+name); err != nil {
		return err
	}

	// Roll back to it, without being interrupted, unless it's released
	released := false
	rollback := func() error {
		ctx := context.WithoutCancel(ctx)
		return errors.Join(exec(ctx, "ROLLBACK TO "+name), exec(ctx, "RELEASE "+name))
	}
	defer func() {
		if !released {
			_ = rollback()
		}
	}()
	err := fn(&conn{pool: c.pool, tx: c.tx, depth: c.depth + 1, bind: c.bind.Copy()})
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

// readerFunc is a function which scans a row
type readerFunc func(sqlite.Row) error

func (fn readerFunc) Scan(row sqlite.Row) error {
	return fn(row)
}
