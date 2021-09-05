package sqlite

import (
	sql "database/sql/driver"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
	driver "github.com/mattn/go-sqlite3"

	// Import namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type txn struct {
	conn  *driver.SQLiteConn
	st    []*driver.SQLiteStmt
	inner bool
}

type prepared struct {
	SQStatement
	p *driver.SQLiteStmt
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (this *txn) Init(conn *driver.SQLiteConn, inner bool) error {
	this.conn = conn
	this.inner = inner
	return nil
}

func (this *txn) Destroy() error {
	var result error

	// Check for opened connection
	if this.conn == nil {
		return ErrInternalAppError.With("Destroy")
	}

	// Cycle through prepared statements to destroy
	for _, st := range this.st {
		if err := st.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Close connection
	if this.inner == false {
		if err := this.conn.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Release resources
	this.conn = nil
	this.st = nil

	// Return success
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *prepared) String() string {
	return this.Query()
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *txn) Query(q SQStatement, args ...interface{}) (SQRows, error) {
	var results sql.Rows

	// Check opened connection
	if this.conn == nil {
		return nil, ErrBadParameter
	}

	// Bound parameters
	values, err := BoundValues(args)
	if err != nil {
		return nil, err
	}

	// Execute prepared or statement
	if v, ok := q.(*prepared); ok {
		results, err = v.p.Query(values)
	} else {
		results, err = this.conn.Query(q.Query(), values)
	}
	if err != nil {
		return nil, err
	}

	// Create a new resultset object
	return NewRows(results.(*driver.SQLiteRows)), nil
}

func (this *txn) Exec(q SQStatement, args ...interface{}) (SQResult, error) {
	var results sql.Result

	// Check opened connection
	if this.conn == nil {
		return SQResult{}, ErrBadParameter
	}

	// Convert arguments
	values, err := BoundValues(args)
	if err != nil {
		return SQResult{}, err
	}

	// Execute prepared or statement
	if v, ok := q.(*prepared); ok {
		results, err = v.p.Exec(values)
	} else {
		results, err = this.conn.Exec(q.Query(), values)
	}
	if err != nil {
		return SQResult{}, err
	}

	// Return results
	if lastInsertID, err := results.LastInsertId(); err != nil {
		return SQResult{}, err
	} else if rowsAffected, err := results.RowsAffected(); err != nil {
		return SQResult{}, err
	} else {
		return SQResult{lastInsertID, uint64(rowsAffected)}, nil
	}
}

func (this *txn) Prepare(v SQStatement) (SQStatement, error) {
	// Return any prepared statements
	if v, ok := v.(*prepared); ok {
		return v, nil
	}
	// Prepare the statement
	if stmt, err := this.conn.Prepare(v.Query()); err != nil {
		return nil, err
	} else {
		return &prepared{v, stmt.(*driver.SQLiteStmt)}, nil
	}
}
