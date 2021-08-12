package sqlite

import (
	sql "database/sql/driver"
	"fmt"
	"reflect"
	"time"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
	multierror "github.com/hashicorp/go-multierror"
	driver "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type txn struct {
	conn  *driver.SQLiteConn
	st    []*driver.SQLiteStmt
	inner bool
}

type prepared struct {
	sqlite.SQStatement
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
		return sqlite.ErrInternalAppError
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

func (this *txn) Query(q sqlite.SQStatement, args ...interface{}) (sqlite.SQRows, error) {
	var results sql.Rows

	// Check opened connection
	if this.conn == nil {
		return nil, sqlite.ErrBadParameter
	}

	// Convert arguments
	values, err := to_values(args)
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

func (this *txn) Exec(q sqlite.SQStatement, args ...interface{}) (sqlite.SQResult, error) {
	var results sql.Result

	// Check opened connection
	if this.conn == nil {
		return sqlite.SQResult{}, sqlite.ErrBadParameter
	}

	// Convert arguments
	values, err := to_values(args)
	if err != nil {
		return sqlite.SQResult{}, err
	}

	// Execute prepared or statement
	if v, ok := q.(*prepared); ok {
		results, err = v.p.Exec(values)
	} else {
		results, err = this.conn.Exec(q.Query(), values)
	}
	if err != nil {
		return sqlite.SQResult{}, err
	}

	// Return results
	if lastInsertID, err := results.LastInsertId(); err != nil {
		return sqlite.SQResult{}, err
	} else if rowsAffected, err := results.RowsAffected(); err != nil {
		return sqlite.SQResult{}, err
	} else {
		return sqlite.SQResult{lastInsertID, uint64(rowsAffected)}, nil
	}
}

func (this *txn) Prepare(v sqlite.SQStatement) (sqlite.SQStatement, error) {
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

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// to_values converts all values to supported types or returns error
func to_values(args []interface{}) ([]sql.Value, error) {
	v := make([]sql.Value, len(args))
	for i, arg := range args {
		// Promote uint and int to int64
		switch arg.(type) {
		case int:
			v[i] = int64(arg.(int))
		case int8:
			v[i] = int64(arg.(int8))
		case int16:
			v[i] = int64(arg.(int16))
		case int32:
			v[i] = int64(arg.(int32))
		case float64:
			v[i] = float64(arg.(float64))
		case float32:
			v[i] = float64(arg.(float32))
		case string, int64, time.Time, bool, nil, []byte:
			v[i] = arg
		default:
			return nil, fmt.Errorf("Unsupported bind type: %v (argument %v)", reflect.TypeOf(arg), i)
		}
	}

	// Return supported values
	return v, nil
}
