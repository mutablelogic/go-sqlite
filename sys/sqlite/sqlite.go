/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package sqlite

import (
	sql "database/sql/driver"
	"fmt"
	"strconv"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	errors "github.com/djthorpe/gopi/util/errors"
	sq "github.com/djthorpe/sqlite"
	driver "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Path string
}

type sqlite struct {
	log        gopi.Logger
	conn       *driver.SQLiteConn
	path       string
	statements []*statement
}

type statement struct {
	query     string
	statement *driver.SQLiteStmt
}

type resultset struct {
	rows    *driver.SQLiteRows
	columns []sq.Column
	values  []sql.Value
}

type column struct {
	name     string
	decltype string
	nullable bool
	pos      int
}

type value struct {
	v sql.Value
	c *column
}

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	sqLiteDriver        = &driver.SQLiteDriver{}
	sqLiteVersion, _, _ = driver.Version()
)

const (
	DEFAULT_COLUMN_TYPE = "TEXT"
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Config) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<sqlite.Open>{ config=%+v version=%v }", config, strconv.Quote(sqLiteVersion))

	this := new(sqlite)
	this.log = logger

	if db, err := sqLiteDriver.Open(config.Path); err != nil {
		return nil, err
	} else {
		this.conn = db.(*driver.SQLiteConn)
		this.path = config.Path
		this.statements = make([]*statement, 0, 10)
	}

	// Success
	return this, nil
}

func (this *sqlite) Close() error {
	this.log.Debug("<sqlite.Close>{ dsn=%v }", strconv.Quote(this.path))

	var err errors.CompoundError

	// Check for opened connection
	if this.conn == nil {
		return gopi.ErrAppError
	}

	// Cycle through prepared statements to destroy
	for _, s := range this.statements {
		if s.statement != nil {
			this.log.Debug2("<sqlite.Destroy>{ %v }", s)
			err.Add(s.statement.Close())
		}
	}

	// Close connection
	err.Add(this.conn.Close())

	// Release resources
	this.statements = nil
	this.conn = nil

	// Return success
	return err.ErrorOrSelf()
}

func (this *sqlite) Version() string {
	return sqLiteVersion
}

func (this *sqlite) Tables() []string {
	if this.conn == nil {
		return nil
	} else if rows, err := this.QueryOnce("SELECT name FROM sqlite_master WHERE type=? ORDER BY name ASC", "table"); err != nil {
		this.log.Error("Tables: %v", err)
		return nil
	} else {
		names := make([]string, 0, 10)
		for {
			values := rows.Next()
			if values == nil {
				break
			} else if len(values) != 1 {
				this.log.Warn("Tables: Expected a single value")
				return nil
			} else {
				names = append(names, values[0].String())
			}
		}
		return names
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqlite) String() string {
	return fmt.Sprintf("<sqlite>{ dsn=%v version=%v }", strconv.Quote(this.path), strconv.Quote(this.Version()))
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqlite) Prepare(query string) (sq.Statement, error) {
	this.log.Debug2("<sqlite.Prepare>{ %v }", strconv.Quote(query))

	if this.conn == nil {
		return nil, gopi.ErrAppError
	} else if prepared, err := this.conn.Prepare(query); err != nil {
		return nil, err
	} else {
		st := &statement{query, prepared.(*driver.SQLiteStmt)}
		this.statements = append(this.statements, st)
		return st, nil
	}
}

func (this *sqlite) Destroy(query sq.Statement) error {
	this.log.Debug2("<sqlite.Destroy>{ %v }", query)

	if this.conn == nil {
		return gopi.ErrAppError
	} else if query_, ok := query.(*statement); ok == false {
		return gopi.ErrBadParameter
	} else {
		var err error
		if query_.statement != nil {
			err = query_.statement.Close()
			query_.statement = nil
		}
		return err
	}
}

func (this *sqlite) Do(query sq.Statement, args ...interface{}) (sq.Result, error) {
	this.log.Debug2("<sqlite.Do>{ %v num_input=%v }", query, len(args))

	if this.conn == nil {
		return sq.Result{}, gopi.ErrAppError
	} else if query == nil {
		return sq.Result{}, gopi.ErrBadParameter
	} else if query_, ok := query.(*statement); ok == false || query_.statement == nil {
		return sq.Result{}, gopi.ErrBadParameter
	} else if query_.statement == nil {
		return sq.Result{}, gopi.ErrBadParameter
	} else if values, err := to_values(args, query_.statement.NumInput()); err != nil {
		return sq.Result{}, err
	} else if results, err := query_.statement.Exec(values); err != nil {
		return sq.Result{}, err
	} else if lastInsertID, err := results.LastInsertId(); err != nil {
		return sq.Result{}, err
	} else if rowsAffected, err := results.RowsAffected(); err != nil {
		return sq.Result{}, err
	} else {
		return sq.Result{lastInsertID, uint64(rowsAffected)}, nil
	}
}

func (this *sqlite) DoOnce(query string, args ...interface{}) (sq.Result, error) {
	this.log.Debug2("<sqlite.DoOnce>{ %v num_input=%v }", strconv.Quote(query), len(args))

	if this.conn == nil {
		return sq.Result{}, gopi.ErrAppError
	} else if values, err := to_values(args, -1); err != nil {
		return sq.Result{}, err
	} else if results, err := this.conn.Exec(query, values); err != nil {
		return sq.Result{}, err
	} else if lastInsertID, err := results.LastInsertId(); err != nil {
		return sq.Result{}, err
	} else if rowsAffected, err := results.RowsAffected(); err != nil {
		return sq.Result{}, err
	} else {
		return sq.Result{lastInsertID, uint64(rowsAffected)}, nil
	}
}

func (this *sqlite) Query(query sq.Statement, args ...interface{}) (sq.Rows, error) {
	this.log.Debug2("<sqlite.Query>{ %v num_input=%v }", query, len(args))

	if this.conn == nil {
		return nil, gopi.ErrAppError
	} else if query == nil {
		return nil, gopi.ErrBadParameter
	} else if query_, ok := query.(*statement); ok == false || query_.statement == nil {
		return nil, gopi.ErrBadParameter
	} else if values, err := to_values(args, query_.statement.NumInput()); err != nil {
		return nil, err
	} else if rows, err := query_.statement.Query(values); err != nil {
		return nil, err
	} else if rs, err := to_rows(rows.(*driver.SQLiteRows)); err != nil {
		return nil, err
	} else {
		// Check columns
		for _, column := range rs.Columns() {
			if sq.IsSupportedType(column.DeclType()) == false {
				this.log.Warn("Warning: Column %v is not a supported type", strconv.Quote(column.Name()))
			}
		}
		// Return resultset
		return rs, nil
	}
}

func (this *sqlite) QueryOnce(query string, args ...interface{}) (sq.Rows, error) {
	this.log.Debug2("<sqlite.QueryOnce>{ %v num_input=%v }", query, len(args))

	if this.conn == nil {
		return nil, gopi.ErrAppError
	} else if values, err := to_values(args, -1); err != nil {
		return nil, err
	} else if rows, err := this.conn.Query(query, values); err != nil {
		return nil, err
	} else if rs, err := to_rows(rows.(*driver.SQLiteRows)); err != nil {
		return nil, err
	} else {
		// Check columns
		for _, column := range rs.Columns() {
			if sq.IsSupportedType(column.DeclType()) == false {
				this.log.Warn("Warning: Column %v is not a supported type", strconv.Quote(column.Name()))
			}
		}
		// Return resultset
		return rs, nil
	}
}

func (this *sqlite) Tx(cb func(sq.Connection) error) error {
	this.log.Debug("<sqlite.Tx>{ BEGIN }")

	if this.conn == nil {
		return gopi.ErrAppError
	} else if tx, err := this.conn.Begin(); err != nil {
		return err
	} else if err := cb(this); err != nil {
		this.log.Debug("<sqlite.Tx>{ ROLLBACK ERROR=%v }", err)
		var errs errors.CompoundError
		errs.Add(err)
		errs.Add(tx.Rollback())
		return errs.ErrorOrSelf()
	} else if err := tx.Commit(); err != nil {
		this.log.Debug("<sqlite.Tx>{ COMMIT ERROR=%v }", err)
		return err
	} else {
		this.log.Debug("<sqlite.Tx>{ COMMIT OK }")
		return nil
	}
}
