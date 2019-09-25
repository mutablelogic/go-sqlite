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
	"strconv"
	"strings"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	errors "github.com/djthorpe/gopi/util/errors"
	sq "github.com/djthorpe/sqlite"
	driver "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type txn struct {
	log   gopi.Logger
	conn  *driver.SQLiteConn
	st    []*driver.SQLiteStmt
	inner bool
	sync.Mutex
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (this *txn) Init(conn *driver.SQLiteConn, inner bool, logger gopi.Logger) error {
	this.conn = conn
	this.log = logger
	this.inner = inner
	this.st = make([]*driver.SQLiteStmt, 0, 10)

	// Success
	return nil
}

func (this *txn) Destroy() error {
	// Close connection
	var err errors.CompoundError

	// Check for opened connection
	if this.conn == nil {
		return gopi.ErrAppError
	}

	// Cycle through prepared statements to destroy
	for _, st := range this.st {
		err.Add(st.Close())
	}

	if this.inner == false {
		err.Add(this.conn.Close())
	}

	// Release resources
	this.conn = nil
	this.st = nil

	// Return success
	return err.ErrorOrSelf()
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *txn) String() string {
	return fmt.Sprintf("<sqtxn>{ conn=%v num_st=%v inner=%v }", this.conn, len(this.st), this.inner)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *txn) NewStatement(query string) sq.Statement {
	this.log.Debug2("<sqtxn.NewStatement>{ %v }", strconv.Quote(query))
	return &statement{prepared{nil}, query}
}

func (this *txn) NewColumn(name, decltype string, nullable, primary bool) sq.Column {
	this.log.Debug2("<sqtxn.NewColumn>{ name=%v decltype=%v nullable=%v primary=%v }", strconv.Quote(name), strconv.Quote(decltype), nullable, primary)

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else if decltype = strings.TrimSpace(decltype); decltype == "" {
		return nil
	} else {
		return &column{name, decltype, nullable, primary, -1}
	}
}

func (this *txn) NewColumnWithIndex(name, decltype string, nullable, primary bool, index int) sq.Column {
	this.log.Debug2("<sqtxn.NewColumn>{ name=%v decltype=%v nullable=%v primary=%v index=%v }", strconv.Quote(name), strconv.Quote(decltype), nullable, primary, index)

	if name = strings.TrimSpace(name); name == "" {
		return nil
	} else if decltype = strings.TrimSpace(decltype); decltype == "" {
		return nil
	} else {
		return &column{name, decltype, nullable, primary, index}
	}
}

func (this *txn) Do(query sq.Statement, args ...interface{}) (sq.Result, error) {
	this.log.Debug2("<sqtxn.Do>{ %v num_input=%v }", strconv.Quote(query.Query()), len(args))

	if this.conn == nil {
		return sq.Result{}, gopi.ErrAppError
	} else if query == nil {
		return sq.Result{}, gopi.ErrBadParameter
	} else if results, err := this.do(query, args); err != nil {
		return sq.Result{}, err
	} else if lastInsertID, err := results.LastInsertId(); err != nil {
		return sq.Result{}, err
	} else if rowsAffected, err := results.RowsAffected(); err != nil {
		return sq.Result{}, err
	} else {
		return sq.Result{lastInsertID, uint64(rowsAffected)}, nil
	}
}

func (this *txn) DoOnce(query string, args ...interface{}) (sq.Result, error) {
	this.log.Debug2("<sqtxn.DoOnce>{ %v num_input=%v }", strconv.Quote(query), len(args))

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

func (this *txn) Query(query sq.Statement, args ...interface{}) (sq.Rows, error) {
	this.log.Debug2("<sqtxn.Query>{ %v num_input=%v }", strconv.Quote(query.Query()), len(args))

	if this.conn == nil {
		return nil, gopi.ErrAppError
	} else if query == nil {
		return nil, gopi.ErrBadParameter
	} else if rows, err := this.query(query, args); err != nil {
		return nil, err
	} else if rs, err := to_rows(rows.(*driver.SQLiteRows)); err != nil {
		return nil, err
	} else {
		// Check columns
		for _, column := range rs.Columns() {
			if sq.IsSupportedType(column.DeclType()) == false {
				this.log.Warn("Warning: Column %v is not a supported type (%v)", strconv.Quote(column.Name()), column.DeclType())
			}
		}
		// Return resultset
		return rs, nil
	}
}

func (this *txn) QueryOnce(query string, args ...interface{}) (sq.Rows, error) {
	this.log.Debug2("<sqtxn.QueryOnce>{ %v num_input=%v }", strconv.Quote(query), len(args))

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

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *txn) prepare(query sq.Statement) (*driver.SQLiteStmt, error) {
	if st, ok := query.(statement_iface); ok == false {
		return nil, gopi.ErrBadParameter
	} else {
		if st.Stmt() == nil {
			if prepared, err := this.conn.Prepare(query.Query()); err != nil {
				return nil, err
			} else if prepared_, ok := prepared.(*driver.SQLiteStmt); ok == false || prepared_ == nil {
				return nil, gopi.ErrAppError
			} else {
				this.st = append(this.st, prepared_)
				st.SetStmt(prepared_)
			}
		}
		return st.Stmt(), nil
	}
}

func (this *txn) do(query sq.Statement, args []interface{}) (sql.Result, error) {
	this.Lock()
	defer this.Unlock()

	if st, err := this.prepare(query); err != nil {
		return &driver.SQLiteResult{}, err
	} else if st == nil {
		return &driver.SQLiteResult{}, gopi.ErrAppError
	} else if values, err := to_values(args, st.NumInput()); err != nil {
		return &driver.SQLiteResult{}, err
	} else if results, err := st.Exec(values); err != nil {
		return &driver.SQLiteResult{}, err
	} else {
		return results, nil
	}
}

func (this *txn) query(query sq.Statement, args []interface{}) (sql.Rows, error) {
	this.Lock()
	defer this.Unlock()

	if st, err := this.prepare(query); err != nil {
		return nil, err
	} else if st == nil {
		return nil, gopi.ErrAppError
	} else if values, err := to_values(args, st.NumInput()); err != nil {
		return nil, err
	} else if rows, err := st.Query(values); err != nil {
		return nil, err
	} else {
		return rows, nil
	}
}
