package sqobj

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"

	// Modules
	marshaler "github.com/djthorpe/go-marshaler"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
	multierror "github.com/hashicorp/go-multierror"

	// Import Namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqclass struct {
	sync.RWMutex
	SQSource
	t reflect.Type

	// All columns, primary key, unique and index column names
	columns []*sqcolumn
	indexes map[string]*sqindex

	// Prepared statements
	s map[SQKey][]SQStatement
}

type sqcolumn struct {
	SQColumn
	*marshaler.Field
	Index   bool
	Unique  bool
	Foreign bool
	Auto    bool
}

type sqindex struct {
	name   string
	unique bool
	cols   []string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	encoder = marshaler.NewEncoder(TagName)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewClass(name, schema string, proto interface{}) *sqclass {
	this := new(sqclass)
	this.s = make(map[SQKey][]SQStatement)
	this.indexes = make(map[string]*sqindex)

	// Set source
	this.SQSource = N(name).WithSchema(schema)

	// Set type - must be a struct
	if v := valueOf(proto); !v.IsValid() {
		return nil
	} else {
		this.t = v.Type()
	}

	// Set column definitions
	fields := encoder.Reflect(proto)
	if len(fields) == 0 {
		return nil
	}
	for _, field := range fields {
		c := &sqcolumn{Field: field}
		c.SQColumn = C(field.Name).WithType(decltype(field.Type))
		for _, tag := range field.Tags {
			if IsSupportedType(tag) {
				c.SQColumn = c.SQColumn.WithType(strings.ToUpper(tag))
			} else if isNotNull(tag) {
				c.SQColumn = c.SQColumn.NotNull()
			} else if isPrimary(tag) {
				c.SQColumn = c.SQColumn.NotNull().WithPrimary()
			} else if isAutoincrement(tag) {
				c.SQColumn = c.SQColumn.WithAutoIncrement().WithPrimary().NotNull()
				c.Auto = true
			} else if isUnique(tag) {
				c.Unique = true
			} else if isIndex(tag) {
				c.Index = true
			} else if isForeign(tag) {
				c.Foreign = true
			}
		}
		this.columns = append(this.columns, c)
	}

	// Get indexes for the prototype
	for _, field := range fields {
		for _, tag := range field.Tags {
			if strings.HasPrefix(tag, "index:") {
				if _, exists := this.indexes[tag]; exists {
					this.indexes[tag].cols = append(this.indexes[tag].cols, field.Name)
				} else {
					this.indexes[tag] = &sqindex{tag[6:], false, []string{field.Name}}
				}
			} else if strings.HasPrefix(tag, "unique:") {
				if _, exists := this.indexes[tag]; exists {
					this.indexes[tag].cols = append(this.indexes[tag].cols, field.Name)
				} else {
					this.indexes[tag] = &sqindex{tag[7:], true, []string{field.Name}}
				}
			}
		}
	}

	// Return success
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
	str += fmt.Sprintf(" columns=%q", this.ColumnNames())
	str += fmt.Sprintf(" primary=%q", this.PrimaryColumnNames())
	str += fmt.Sprintf(" foreign=%q", this.ForeignColumnNames())
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// Return the type for the class
func (this *sqclass) Type() reflect.Type {
	return this.t
}

// Return a new empty object (pointer to element)
func (this *sqclass) Proto() interface{} {
	return reflect.New(this.t).Interface()
}

// Get SQL prepared statements
func (this *sqclass) Get(k SQKey) []SQStatement {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()

	if st, exists := this.s[k]; exists {
		return st
	} else {
		return nil
	}
}

// Set SQL statements
func (this *sqclass) Set(v ...SQStatement) SQKey {
	this.RWMutex.Lock()
	defer this.RWMutex.Unlock()

	// Check parameters
	if len(v) == 0 {
		return 0
	}

	// Create a new key and add the statements
	k := this.newkey()
	this.s[k] = v
	return k
}

// Columns returns an array of column definitions
func (this *sqclass) Columns() []SQColumn {
	result := make([]SQColumn, 0, len(this.columns))
	for _, col := range this.columns {
		result = append(result, col.SQColumn)
	}
	return result
}

// ColumnNames returns an array of column names
func (this *sqclass) ColumnNames() []string {
	result := make([]string, 0, len(this.columns))
	for _, col := range this.columns {
		result = append(result, col.Field.Name)
	}
	return result
}

// ColumnSources returns an array of column sources
func (this *sqclass) ColumnSources() []SQSource {
	result := make([]SQSource, 0, len(this.columns))
	for _, col := range this.columns {
		result = append(result, N(col.Field.Name))
	}
	return result
}

// PrimaryColumnNames returns an array of column names which are
// used in the primary key.
func (this *sqclass) PrimaryColumnNames() []string {
	result := make([]string, 0, len(this.columns))
	for _, col := range this.columns {
		if col.Primary() != "" {
			result = append(result, col.Field.Name)
		}
	}
	return result
}

// ForeignColumnNames returns an array of columns which are included
// in the default foreign key
func (this *sqclass) ForeignColumnNames() []string {
	result := make([]string, 0, len(this.columns))
	for _, col := range this.columns {
		if col.Foreign {
			result = append(result, col.Field.Name)
		}
	}
	return result
}

func (this *sqclass) WithForeignKey(class SQClass, columns ...string) error {
	foreign := this.ForeignColumnNames()
	if class == nil || class.Name() == this.Name() {
		// no self-references
		return ErrBadParameter.Withf("WithForeignKey: %q", this.Name())
	} else if len(foreign) == 0 {
		// undefined column
		return ErrBadParameter.Withf("WithForeignKey: %q: No defined foreign keys", this.Name())
	} else if len(foreign) != len(columns) && len(columns) != 0 {
		// Bad number of parameters
		return ErrBadParameter.Withf("WithForeignKey: %q: Column mis-match", this.Name())
	} else if st, exists := this.s[SQKeyCreate]; !exists || len(st) == 0 {
		// Create statement does not yet exist
		return ErrInternalAppError.Withf("WithForeignKey: %q", this.Name())
	} else {
		// Add foreign key to existing create statement
		this.RWMutex.Lock()
		defer this.RWMutex.Unlock()
		st[0] = st[0].(SQTable).WithForeignKey(class.ForeignKey(columns...).OnDeleteCascade(), foreign...)
	}

	// Return success
	return nil
}

// NewIterator creates a read iterator for a resultset
func (this *sqclass) NewIterator(rs SQRows) SQIterator {
	return NewIterator(this, rs)
}

// Values returns the parameters for an insert statement
// Sets values in autoincrement fields to NULL if the value in
// the autoincrement field is a zero value.
func (this *sqclass) Values(v interface{}) ([]interface{}, error) {
	var errs error

	// Check parameter
	rv := valueOf(v)
	if !rv.IsValid() || rv.Type() != this.t {
		return nil, ErrBadParameter.Withf("Value: %q", rv.Type())
	}

	// Iterate over columns
	result := make([]interface{}, 0, len(this.columns))
	for _, col := range this.columns {
		value := rv.Field(col.Field.Index)
		if col.Auto && value.IsZero() {
			result = append(result, nil)
		} else if v, err := sqlite.BoundValue(value); err != nil {
			err = multierror.Append(errs, err)
		} else {
			result = append(result, v)
		}
	}

	// Return result
	return result, errs
}

// Object returns a new object from a select statement row
func (this *sqclass) Object(v []interface{}) (interface{}, error) {
	var errs error

	// Make an object
	proto := reflect.New(this.t)

	// Iterate over columns
	for i, col := range this.columns {
		if value, err := sqlite.UnboundValue(v[i], col.Field.Type); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			proto.Elem().Field(col.Field.Index).Set(value)
		}
	}

	// Return result
	return proto.Interface(), errs
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - STATEMENTS

func (this *sqclass) newkey() SQKey {
	k := SQKey(rand.Uint32())
	if k < SQKeyMax {
		return this.newkey()
	} else if _, exists := this.s[k]; exists {
		return this.newkey()
	} else {
		return k
	}
}

// return primary key parameters or nil on error
func (this *sqclass) primaryvalues(v interface{}) ([]interface{}, error) {
	var result []interface{}
	var err error
	r := reflect.ValueOf(v)
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}
	if r.Kind() != reflect.Struct {
		return nil, ErrBadParameter
	}
	for _, field := range this.columns {
		if field.Primary() != "" {
			if v, err_ := sqlite.BoundValue(r.Field(field.Field.Index)); err != nil {
				err = multierror.Append(err, err_)
			} else {
				result = append(result, v)
			}
		}
	}
	// Return any errors
	return result, err
}
