/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	sql "database/sql/driver"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	errors "github.com/djthorpe/gopi/util/errors"
	sq "github.com/djthorpe/sqlite"
	driver "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Database struct {
	Path     string
	Location string
}

type sqlite struct {
	log gopi.Logger
	dsn string
	tz  *time.Location
	ctx *driver.SQLiteTx

	txn
	sync.Mutex
}

type prepared struct {
	*driver.SQLiteStmt
}

type statement struct {
	prepared

	query string
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
	primary  bool
	index    int
}

type value struct {
	v sql.Value
	c *column
}

type statement_iface interface {
	// Get prepared statement
	Stmt() *driver.SQLiteStmt

	// Set prepared statement
	SetStmt(*driver.SQLiteStmt)
}

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	sqLiteDriver        = &driver.SQLiteDriver{}
	sqLiteVersion, _, _ = driver.Version()
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Database) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<sqlite.Open>{ config=%+v version=%v }", config, strconv.Quote(sqLiteVersion))

	this := new(sqlite)
	this.log = logger

	if config.Location == "" {
		this.tz = time.Local
	} else if location, err := time.LoadLocation(config.Location); err != nil {
		return nil, err
	} else {
		this.tz = location
	}

	if config.Path == "" {
		config.Path = ":memory:"
	}

	if dsn, err := url.Parse("file:" + config.Path); err != nil {
		return nil, err
	} else {
		q := dsn.Query()
		q.Set("_loc", this.tz.String())
		dsn.RawQuery = q.Encode()
		this.dsn = dsn.String()
		if db, err := sqLiteDriver.Open(this.dsn); err != nil {
			return nil, err
		} else if err := this.txn.Init(db.(*driver.SQLiteConn), false, logger); err != nil {
			return nil, err
		}
	}

	// Success
	return this, nil
}

func (this *sqlite) Close() error {
	this.log.Debug("<sqlite.Close>{ dsn=%v }", strconv.Quote(this.dsn))
	return this.txn.Destroy()
}

func (this *sqlite) Version() string {
	return sqLiteVersion
}

func (this *sqlite) Schemas() []string {
	this.log.Debug2("<sqlite.Schemas>{ }")

	if rows, err := this.QueryOnce("PRAGMA database_list"); err != nil {
		this.log.Error("Schemas: %v", err)
		return nil
	} else {
		schemas := make([]string, 0, 1)
		for {
			row := sq.RowMap(rows.Next())
			if row == nil {
				break
			} else if name, exists := row["name"]; exists {
				schemas = append(schemas, name.String())
			}
		}
		return schemas
	}
}

func (this *sqlite) Tables() []string {
	return this.TablesEx("", false)
}

func (this *sqlite) TablesEx(schema string, temp bool) []string {
	this.log.Debug2("<sqlite.TablesEx>{ schema=%v include_temporary=%v }", strconv.Quote(schema), temp)

	// Create the query
	query := ""
	if temp {
		query = `
			SELECT name FROM 
   				(SELECT name,type FROM %ssqlite_master UNION ALL SELECT name,type FROM %ssqlite_temp_master)
			WHERE type=? AND name NOT LIKE 'sqlite_%%'
			ORDER BY name ASC
		`
	} else {
		query = `
			SELECT name FROM 
				%ssqlite_master 
			WHERE type=? AND name NOT LIKE 'sqlite_%%'
			ORDER BY name ASC -- %s
		`
	}

	// Append the schema
	if schema != "" {
		query = fmt.Sprintf(query, sq.QuoteIdentifier(schema)+".", sq.QuoteIdentifier(schema)+".")
	} else {
		query = fmt.Sprintf(query, "", "")
	}

	// Perform the query
	if rows, err := this.QueryOnce(query, "table"); err != nil {
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
// TABLE INFO

func (this *sqlite) ColumnsForTable(name, schema string) ([]sq.Column, error) {
	this.log.Debug2("<sqlite.ColumnsForTable>{ name=%v schema=%v }", strconv.Quote(name), strconv.Quote(schema))

	if name = strings.TrimSpace(name); name == "" {
		return nil, gopi.ErrBadParameter
	}

	query := "table_info(" + sq.QuoteIdentifier(name) + ")"
	if schema != "" {
		query = "PRAGMA " + sq.QuoteIdentifier(schema) + "." + query
	} else {
		query = "PRAGMA" + query
	}
	if rows, err := this.QueryOnce(query); err != nil {
		return nil, err
	} else {
		columns := make([]sq.Column, 0, 10)
		for {
			row := sq.RowMap(rows.Next())
			if row == nil {
				break
			} else {
				c := &column{
					name:     row["name"].String(),
					decltype: row["type"].String(),
					nullable: row["notnull"].Bool() == false,
				}
				columns = append(columns, c)
			}
		}
		return columns, nil
	}
}

func (this *sqlite) Attach(schema, dsn string) error {
	this.log.Debug2("<sqlite.Attach>{ schema=%v dsn=%v }", strconv.Quote(schema), strconv.Quote(dsn))

	schema = strings.TrimSpace(schema)
	if schema == "" || dsn == "" {
		return gopi.ErrBadParameter
	} else {
		query := "ATTACH DATABASE " + sq.DoubleQuote(dsn) + " AS " + sq.QuoteIdentifier(schema)
		if _, err := this.DoOnce(query); err != nil {
			return err
		}
	}
	// Success
	return nil
}

func (this *sqlite) Detach(schema string) error {
	this.log.Debug2("<sqlite.Detach>{ schema=%v }", strconv.Quote(schema))

	schema = strings.TrimSpace(schema)
	if schema == "" {
		return gopi.ErrBadParameter
	} else {
		query := "DETACH DATABASE " + sq.QuoteIdentifier(schema)
		if _, err := this.DoOnce(query); err != nil {
			return err
		}
	}
	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqlite) String() string {
	return fmt.Sprintf("<sqlite>{ dsn=%v tz=%v version=%v }", strconv.Quote(this.dsn), this.tz, strconv.Quote(this.Version()))
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqlite) Txn(cb func(sq.Transaction) error) error {
	this.log.Debug("<sqlite.Txn>{ BEGIN }")
	var errs errors.CompoundError

	transaction := new(txn)
	if this.ctx != nil {
		return gopi.ErrOutOfOrder
	} else if ctx, err := this.conn.Begin(); err != nil {
		return err
	} else if this.ctx = ctx.(*driver.SQLiteTx); this.ctx == nil {
		return gopi.ErrOutOfOrder
	} else if err := transaction.Init(this.conn, true, this.log); err != nil {
		return err
	}

	this.Lock()
	defer this.Unlock()

	if err := cb(transaction); err != nil {
		this.log.Debug("<sqlite.Tx>{ ROLLBACK ERROR=%v }", err)
		errs.Add(err)
		errs.Add(this.ctx.Rollback())
		this.ctx = nil
		return errs.ErrorOrSelf()
	} else if err := this.ctx.Commit(); err != nil {
		this.log.Debug("<sqlite.Tx>{ COMMIT ERROR=%v }", err)
		this.ctx = nil
		return err
	} else {
		this.log.Debug("<sqlite.Tx>{ COMMIT OK }")
		this.ctx = nil
		return nil
	}
}
