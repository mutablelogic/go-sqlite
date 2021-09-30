package sqobj

import (
	"fmt"
	"reflect"

	// Import Namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Class struct {
	*SQReflect
	SQSource

	// Prepared statements and in-place parameters
	s map[stkey]SQStatement
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
	this.s = make(map[stkey]SQStatement)

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
func (this *Class) Create(txn SQTransaction, schema string) error {
	// If schema then set it
	if schema != "" {
		this.SQSource = this.SQSource.WithSchema(schema)
	}

	if txn.Flags().Is(SQLITE_OPEN_OVERWRITE) && hasElement(txn.Tables(this.Schema()), this.Name()) {
		// Drop indexes
		for _, index := range txn.IndexesForTable(this.Schema(), this.Name()) {
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
		if st := st(this, txn); st == nil {
			return ErrBadParameter.Withf("Create %q: %q", this.Name(), key)
		} else {
			this.s[key] = st
		}
	}

	// Return success
	return nil
}

// Insert into a table and return rowids. If any autoincremented fields are zero valued, these are automatically
// set to NULL on insert
func (c *Class) Insert(txn SQTransaction, v ...interface{}) ([]int64, error) {
	result := make([]int64, 0, len(v))

	// Retrieve prepared statement
	st, exists := c.s[SQKeyInsert]
	if !exists {
		return nil, ErrOutOfOrder.Withf("Insert: %q", c.Name())
	}

	// Insert each object
	for _, v := range v {
		rv := ValueOf(v)
		if !rv.IsValid() || rv.Type() != c.t {
			return nil, ErrBadParameter.Withf("Insert: %v", v)
		}
		r, err := txn.Query(st, c.boundValues(rv, true, false)...)
		if err != nil {
			return nil, err
		}
		result = append(result, r.LastInsertId())
	}

	// Return success
	return result, nil
}

// Read from table and return an iterator. It is expected that Read would
// accept a query, including: order, limit, offset, distinct and a
// list of expressions
func (this *Class) Read(txn SQTransaction) (SQIterator, error) {
	// Retrieve prepared statement
	st, exists := this.s[SQKeySelect]
	if !exists {
		return nil, ErrOutOfOrder.Withf("Read: %q", this.Name())
	}

	// Do query
	rs, err := txn.Query(st)
	if err != nil {
		return nil, err
	} else {
		return iterator(this, rs), nil
	}
}

// Delete from the table based on rowids, returns the number of changes
// made
func (c *Class) DeleteRows(txn SQTransaction, row []int64) (int, error) {
	// Retrieve prepared statement
	st, exists := c.s[SQKeyDeleteRows]
	if !exists {
		return 0, ErrOutOfOrder.Withf("DeleteRows: %q", c.Name())
	}

	// Delete each row
	var n int
	for _, rowid := range row {
		r, err := txn.Query(st, rowid)
		if err != nil {
			return 0, err
		}
		n += r.RowsAffected()
	}

	// Return success
	return n, nil
}

// Delete keys in table based on primary keys. Returns number of deleted rows
func (c *Class) DeleteKeys(txn SQTransaction, v ...interface{}) (int, error) {
	// Retrieve prepared statement
	st, exists := c.s[SQKeyDeleteKeys]
	if !exists {
		return 0, ErrOutOfOrder.Withf("DeleteKeys: %q", c.Name())
	}

	// Delete each object
	var n int
	for _, v := range v {
		rv := ValueOf(v)
		if !rv.IsValid() || rv.Type() != c.t {
			return 0, ErrBadParameter.Withf("DeleteKeys: %v", v)
		}
		r, err := txn.Query(st, c.boundKeys(rv)...)
		if err != nil {
			return 0, err
		}
		n += r.RowsAffected()
	}

	// Return success
	return n, nil
}

// Update objects by primary key, return number of updated rows
func (c *Class) UpdateKeys(txn SQTransaction, v ...interface{}) (int, error) {
	// Retrieve prepared statement
	st, exists := c.s[SQKeyUpdateKeys]
	if !exists {
		return 0, ErrOutOfOrder.Withf("UpdateKeys: %q", c.Name())
	}

	// Update each object
	var n int
	for _, v := range v {
		rv := ValueOf(v)
		if !rv.IsValid() || rv.Type() != c.t {
			return 0, ErrBadParameter.Withf("UpdateKeys: %v", v)
		}
		r, err := txn.Query(st, c.boundValues(rv, false, true)...)
		if err != nil {
			return 0, err
		}
		n += r.RowsAffected()
	}

	// Return success
	return n, nil
}

func (c *Class) UpsertKeys(txn SQTransaction, v ...interface{}) ([]int64, error) {
	result := make([]int64, 0, len(v))

	// Retrieve prepared statement
	st, exists := c.s[SQKeyUpsertKeys]
	if !exists {
		return nil, ErrOutOfOrder.Withf("UpdateKeys: %q", c.Name())
	}

	// Update each object
	for _, v := range v {
		rv := ValueOf(v)
		if !rv.IsValid() || rv.Type() != c.t {
			return nil, ErrBadParameter.Withf("UpdateKeys: %v", v)
		}
		r, err := txn.Query(st, c.boundValues(rv, true, false)...)
		if err != nil {
			return nil, err
		}
		if r.RowsAffected() > 0 {
			if r.LastInsertId() == 0 {
				fmt.Println("TODO: Set last insert id as rows affected (was an update)")
				result = append(result, -1)
			} else {
				result = append(result, r.LastInsertId())
			}
		} else {
			result = append(result, 0)
		}
	}

	// Return success
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// boundValues returns sqlite-compatible values for a struct value. If autonull
// argument is true, then any zero-value column is set to NULL. This is so inserts
// can be performed. If primarylast is true, then primary values are put behind non-
// primary values.
func (this *Class) boundValues(v reflect.Value, autonull bool, primarylast bool) []interface{} {
	// Set length of parameters
	this.p = this.p[:len(this.col)]

	// First iteration sets values
	j := 0
	if primarylast {
		for _, col := range this.col {
			field := v.Field(col.Field.Index)
			if !col.Primary {
				this.p[j] = field.Interface()
				j++
			}
		}
	}

	for _, col := range this.col {
		field := v.Field(col.Field.Index)
		if primarylast && !col.Primary {
			continue
		}
		if autonull && col.Auto && field.IsZero() {
			this.p[j] = nil
		} else {
			this.p[j] = field.Interface()
		}
		j++
	}

	// Return success
	return this.p
}

// boundKeys returns sqlite-compatible primary keys for a struct value.
func (this *Class) boundKeys(v reflect.Value) []interface{} {
	// Set length of parameters
	this.p = this.p[:0]

	// Iterate over columns
	for _, col := range this.col {
		field := v.Field(col.Field.Index)
		if col.Primary {
			this.p = append(this.p, field.Interface())
		}
	}

	// Return success
	return this.p
}

// unboundValues fills prototype with values from v. The proto is expected to be
// a pointer to a struct value
func (this *Class) unboundValues(proto reflect.Value, v []interface{}) {
	for i, col := range this.col {
		field := proto.Elem().Field(col.Field.Index)
		field.Set(reflect.ValueOf(v[i]))
	}
}
