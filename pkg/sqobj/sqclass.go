package sqobj

import (
	"fmt"
	"reflect"
	"sync"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
	multierror "github.com/hashicorp/go-multierror"

	// Import Namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqclass struct {
	sync.RWMutex
	*SQReflect
	SQSource

	// Prepared statements and in-place parameters
	s map[SQKey]SQStatement
	p []interface{}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// MustRegisterClass registers a SQObject class, panics if an error
// occurs.
func MustRegisterClass(source SQSource, proto interface{}) SQClass {
	if cls, err := RegisterClass(source, proto); err != nil {
		panic(err)
	} else {
		return cls
	}
}

// RegisterClass registers a SQObject class, returns the class and
// any errors
func RegisterClass(source SQSource, proto interface{}) (*sqclass, error) {
	this := new(sqclass)
	this.s = make(map[SQKey]SQStatement)

	// Check name
	if source.Name() == "" {
		return nil, ErrBadParameter.Withf("source")
	} else {
		this.SQSource = source
	}

	// Do reflection
	if r, err := NewReflect(proto); err != nil {
		return nil, err
	} else {
		this.SQReflect = r
	}

	// Set parameters - used by boundValues to fill in parameters
	this.p = make([]interface{}, 0, len(this.col)+1)

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqclass) String() string {
	str := "<sqclass"
	str += fmt.Sprintf(" name=%q", this.Name())
	if schema := this.Schema(); schema != "" {
		str += fmt.Sprintf(" schema=%q", this.Schema())
	}
	str += " " + fmt.Sprint(this.SQReflect)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// Proto returns a prototype of the class
func (this *sqclass) Proto() reflect.Value {
	return reflect.New(this.t)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ForeignKey appends a foreign key constraint, panics on error. Optionally
// sets the columns to refer to in the parent.
func (this *sqclass) ForeignKey(parent SQClass, parentcols ...string) SQClass {
	if err := this.WithForeignKey(parent, parentcols...); err != nil {
		panic(err)
	}
	return this
}

// WithForeignKey appends a foreign key constraint to the class, returns error.
// Optionally sets the columns to refer to in the parent.
func (this *sqclass) WithForeignKey(parent SQClass, parentcols ...string) error {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()
	if parent, ok := parent.(*sqclass); ok {
		return this.SQReflect.WithForeignKey(parent.SQSource, parentcols...)
	} else {
		return ErrInternalAppError
	}
}

// Create creates a table, keys and prepared statements within a transaction. If
// the flag SQLITE_FLAG_DELETEIFEXISTS is set, then tables and indexes are dropped
// first.
func (this *sqclass) Create(conn SQConnection, schema string, flags SQFlag) error {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	// If schema then set it
	if schema != "" {
		this.SQSource = this.SQSource.WithSchema(schema)
	}

	// Suspend foreign key constraints whilst dropping and creating
	if err := conn.SetForeignKeyConstraints(false); err != nil {
		return err
	}
	defer conn.SetForeignKeyConstraints(true)

	return conn.Do(func(txn SQTransaction) error {
		// Drop table if it exists
		if hasElement(conn.TablesEx(this.Schema(), false), this.Name()) && flags&SQLITE_FLAG_DELETEIFEXISTS != 0 {
			// Drop indexes
			for _, index := range conn.IndexesEx(this.Name(), this.Schema()) {
				if !index.Auto() {
					if _, err := txn.Exec(index.DropIndex()); err != nil {
						return err
					}
				}
			}
			// Drop table
			if _, err := txn.Exec(this.DropTable()); err != nil {
				return err
			}
		}

		// Create tables if they don't exist
		for _, st := range this.Table(this.SQSource, true) {
			if _, err := txn.Exec(st); err != nil {
				return err
			}
		}

		// Prepare statements
		for key, fn := range statements {
			st := fn(this, flags)
			if st == nil {
				return ErrBadParameter.Withf("Prepare: %q", key)
			}
			if st == nil {
				continue
			}
			if st, err := conn.Prepare(st); err != nil {
				return ErrUnexpectedResponse.Withf("Prepare: %v: %v", key, err)
			} else {
				this.s[key] = st
			}
			fmt.Println(st)
		}

		// Return success
		return nil
	})
}

// Read from table and return an iterator. It is expected that Read would
// accept a query, including: order, limit, offset, distinct and a
// list of expressions
func (this *sqclass) Read(conn SQConnection) (SQIterator, error) {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()

	// Read from database
	st, exists := this.s[SQKeySelect]
	if !exists {
		return nil, ErrOutOfOrder.Withf("Read: %q", this.Name())
	}
	rs, err := conn.Query(st)
	if err != nil {
		return nil, err
	} else {
		return NewIterator(this, rs), nil
	}
}

// Insert into a table and return an SQResult. If any autoincremented
// fields are zero valued, these are automatically set to NULL on insert
func (this *sqclass) Insert(conn SQConnection, v ...interface{}) ([]SQResult, error) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	result := make([]SQResult, 0, len(v))
	if err := conn.Do(func(txn SQTransaction) error {
		// Retrieve prepared statement
		st, exists := this.s[SQKeyInsert]
		if !exists {
			return ErrOutOfOrder.Withf("Insert: %q", this.Name())
		}
		for _, v := range v {
			rv := ValueOf(v)
			if !rv.IsValid() || rv.Type() != this.t {
				return ErrBadParameter.Withf("Insert: %v", v)
			}
			params, err := this.boundValues(rv, true)
			if err != nil {
				return err
			}
			r, err := conn.Exec(st, params...)
			if err != nil {
				return err
			}
			result = append(result, r)
		}
		return nil
	}); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

// Delete from a table and return an SQResult. Returns error if no primary keys
// defined on the table
func (this *sqclass) Delete(conn SQConnection, v ...interface{}) ([]SQResult, error) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	return nil, ErrNotImplemented
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// boundValues returns sqlite-compatible values for a struct value. If autonull
// argument is true, then any zero-value column is set to NULL. This is so inserts
// can be performed.
func (this *sqclass) boundValues(v reflect.Value, autonull bool) ([]interface{}, error) {
	var errs error

	// Set length of parameters
	this.p = this.p[:len(this.col)]

	// Iterate over columns
	for i, col := range this.col {
		field := v.Field(col.Field.Index)
		if autonull && col.Auto && field.IsZero() {
			this.p[i] = nil
		} else if v, err := sqlite.BoundValue(field); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			this.p[i] = v
		}
	}

	// Return success
	return this.p, nil
}

// unboundValues fills prototype with values from v. The proto is expected to be
// a pointer to a struct value
func (this *sqclass) unboundValues(proto reflect.Value, v []interface{}) error {
	var errs error

	// Iterate over columns
	for i, col := range this.col {
		field := proto.Elem().Field(col.Field.Index)
		if value, err := sqlite.UnboundValue(v[i], field.Type()); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			field.Set(value)
		}
	}

	// Return any errors
	return errs
}
