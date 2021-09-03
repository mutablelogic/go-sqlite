package sqlite

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	// Modules
	multierror "github.com/hashicorp/go-multierror"
	driver "github.com/mattn/go-sqlite3"

	// Import namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type connection struct {
	sync.Mutex
	tz  *time.Location
	dsn string
	ctx *driver.SQLiteTx
	txn
}

///////////////////////////////////////////////////////////////////////////////
// NEW

// Open a database and set the timezone
func Open(path string, location *time.Location) (SQConnection, error) {
	this := new(connection)

	// Set timezone
	if location == nil {
		this.tz = time.Local
	} else {
		this.tz = location
	}

	// Set path
	if path == "" {
		path = sqLiteMemory
	}

	// Open database
	if dsn, err := url.Parse("file:" + path); err != nil {
		return nil, err
	} else {
		q := dsn.Query()
		q.Set("_loc", this.tz.String())
		dsn.RawQuery = q.Encode()
		this.dsn = dsn.String()
		if db, err := sqLiteDriver.Open(this.dsn); err != nil {
			return nil, err
		} else if err := this.txn.Init(db.(*driver.SQLiteConn), false); err != nil {
			return nil, err
		}
	}

	// Return success
	return this, nil
}

// Create an in-memory database and optionally set the timezone
func New(location ...*time.Location) (SQConnection, error) {
	if len(location) == 0 {
		return Open("", nil)
	} else if len(location) == 1 {
		return Open("", location[0])
	} else {
		return nil, ErrBadParameter
	}
}

// Close the database
func (this *connection) Close() error {
	return this.txn.Destroy()
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *connection) String() string {
	str := "<sqlite.connection"
	str += fmt.Sprintf(" dsn=%q", this.dsn)
	if schemas := this.Schemas(); len(schemas) > 0 {
		str += fmt.Sprintf(" schemas=%q", strings.Join(schemas, ","))
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *connection) Do(cb func(SQTransaction) error) error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	transaction := new(txn)
	if this.ctx != nil {
		return ErrInternalAppError.With("Already in a transaction")
	} else if ctx, err := this.conn.Begin(); err != nil {
		return err
	} else if this.ctx = ctx.(*driver.SQLiteTx); this.ctx == nil {
		return ErrInternalAppError.With("Invalid transaction object")
	} else if err := transaction.Init(this.conn, true); err != nil {
		return err
	}

	// Clear transaction after rollback or commit
	defer func() {
		this.ctx = nil
	}()

	// Perform the transaction, on error rollback
	if err := cb(transaction); err != nil {
		var result error
		result = multierror.Append(result, err)
		if err := this.ctx.Rollback(); err != nil {
			result = multierror.Append(result, err)
		}
		return result
	}

	return this.ctx.Commit()
}
