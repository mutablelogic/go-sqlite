package sqobj

import (
	"context"
	"fmt"
	"reflect"

	// Import Namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
	// Packages
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Class struct {
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
func MustRegisterClass(source SQSource, proto interface{}) *Class {
	if cls, err := RegisterClass(source, proto); err != nil {
		panic(err)
	} else {
		return cls
	}
}

// RegisterClass registers a SQObject class, returns the class and
// any errors
func RegisterClass(source SQSource, proto interface{}) (*Class, error) {
	this := new(Class)
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

func (this *Class) String() string {
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
func (this *Class) Proto() reflect.Value {
	return reflect.New(this.t)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ForeignKey appends a foreign key constraint, panics on error. Optionally
// sets the columns to refer to in the parent.
func (this *Class) ForeignKey(parent SQClass, parentcols ...string) SQClass {
	if err := this.WithForeignKey(parent, parentcols...); err != nil {
		panic(err)
	}
	return this
}

// WithForeignKey appends a foreign key constraint to the class, returns error.
// Optionally sets the columns to refer to in the parent.
func (this *Class) WithForeignKey(parent SQClass, parentcols ...string) error {
	if parent, ok := parent.(*Class); ok {
		return this.SQReflect.WithForeignKey(parent.SQSource, parentcols...)
	} else {
		return ErrInternalAppError
	}
}

// Create creates a table, keys and prepared statements within a transaction. If
// the flag SQLITE_OPEN_OVERWRITE is set when creating the connection, then tables
// and indexes are dropped and then re-created.
func (this *Class) Create(ctx context.Context, conn SQConnection, schema string) error {
	// If schema then set it
	if schema != "" {
		this.SQSource = this.SQSource.WithSchema(schema)
	}

	// Do transaction without foreign keys
	return conn.Do(ctx, SQLITE_TXN_NO_FOREIGNKEY_CONSTRAINTS, func(txn SQTransaction) error {
		// Drop table and indexes if it exists
		if conn.Flags().Is(SQLITE_OPEN_OVERWRITE) && hasElement(conn.Tables(this.Schema()), this.Name()) {
			// Drop indexes
			for _, index := range conn.IndexesForTable(this.Schema(), this.Name()) {
				if !index.Auto() {
					if _, err := txn.Query(index.DropIndex()); err != nil {
						return err
					}
				}
			}
			// Drop table
			if _, err := txn.Query(this.DropTable()); err != nil {
				return err
			}
		}

		// Create tables if they don't exist
		for _, st := range this.Table(this.SQSource, true) {
			if _, err := txn.Query(st); err != nil {
				return err
			}
		}

		// Prepare statements for insert, update and delete for example
		for key, st := range statements {
			this.s[key] = st(this, conn)
		}

		// Return success
		return nil
	})
}

// Insert into a table and return rowids. If any autoincremented fields are zero valued, these are automatically
// set to NULL on insert
func (this *Class) Insert(txn SQTransaction, v ...interface{}) ([]int64, error) {
	result := make([]int64, 0, len(v))

	// Retrieve prepared statement
	st, exists := this.s[SQKeyInsert]
	if !exists {
		return nil, ErrOutOfOrder.Withf("Insert: %q", this.Name())
	}
	// Insert each object
	for _, v := range v {
		rv := ValueOf(v)
		if !rv.IsValid() || rv.Type() != this.t {
			return nil, ErrBadParameter.Withf("Insert: %v", v)
		}
		r, err := txn.Query(st, this.boundValues(rv, true)...)
		if err != nil {
			return nil, err
		}
		result = append(result, r.LastInsertId())
	}

	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// boundValues returns sqlite-compatible values for a struct value. If autonull
// argument is true, then any zero-value column is set to NULL. This is so inserts
// can be performed.
func (this *Class) boundValues(v reflect.Value, autonull bool) []interface{} {
	// Set length of parameters
	this.p = this.p[:len(this.col)]

	// Iterate over columns
	for i, col := range this.col {
		field := v.Field(col.Field.Index)
		if autonull && col.Auto && field.IsZero() {
			this.p[i] = nil
		} else {
			this.p[i] = field.Interface()
		}
	}

	// Return success
	return this.p
}

/*
// Read from table and return an iterator. It is expected that Read would
// accept a query, including: order, limit, offset, distinct and a
// list of expressions
func (this *Class) Read(conn SQConnection) (SQIterator, error) {
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

/*


// Delete from a table and return an SQResult. Returns error if no primary keys
// defined on the table
func (this *sqclass) Delete(conn SQConnection, v ...interface{}) ([]SQResult, error) {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()
	return nil, ErrNotImplemented
}
*/
/*
// unboundValues fills prototype with values from v. The proto is expected to be
// a pointer to a struct value
func (this *Class) unboundValues(proto reflect.Value, v []interface{}) error {
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
*/
