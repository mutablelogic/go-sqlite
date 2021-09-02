package sqobj

import (
	"fmt"
	"reflect"

	// Modules
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type classKey uint

type sqclass struct {
	SQSource
	t reflect.Type
	s map[classKey]SQStatement
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	classKeyNone classKey = iota
	classKeyWrite
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewClass(name, schema string, proto interface{}) *sqclass {
	this := new(sqclass)
	this.s = make(map[classKey]SQStatement)

	// Set source
	this.SQSource = N(name).WithSchema(schema)

	// Set type - must be a struct
	v := reflect.ValueOf(proto)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	} else {
		this.t = v.Type()
	}

	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqclass) String() string {
	str := "<sqclass"
	if name := this.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if t := this.Type(); t != nil {
		str += fmt.Sprintf(" type=%q", t)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *sqclass) Type() reflect.Type {
	return this.t
}

func (this *sqclass) Proto() interface{} {
	return reflect.New(this.t).Interface()
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - INSERT

// write returns an INSERT statement for the given struct or nil if the
// argument is not a pointer to a struct or has no fields which are exported
func (this *sqclass) write() SQStatement {
	cols, _, _ := structCols(this.Proto())
	if len(cols) == 0 {
		return nil
	}
	// Conflict occurs on primary key on insert
	// There may be other constraints not checked for, but these should fail as
	// usual
	pk := primaryForColumns(cols)
	if len(pk) == 0 {
		pk = []string{"rowid"}
	}
	// Return the insert statement
	return this.Insert(namesForColumns(cols)...).WithConflictUpdate(pk...)
}

// prepare all statements within a transaction
func (this *sqclass) prepare(txn SQTransaction) error {
	// TODO
	this.s[classKeyWrite] = this.write()

	// Return success
	return nil
}

// return prepared statement
func (this *sqclass) statement(key classKey) SQStatement {
	if st, exists := this.s[key]; exists {
		return st
	} else {
		return nil
	}
}

// return bound parameters
func (this *sqclass) params(v interface{}) ([]interface{}, error) {
	if reflect.TypeOf(v) != this.t {
		return nil, ErrBadParameter.With(reflect.TypeOf(v))
	} else {
		return InsertParams(v)
	}
}
