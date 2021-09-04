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
	v := reflect.ValueOf(proto)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
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

	// Create statments for create,insert and delete
	this.addStatement(SQKeyCreate, this.sqCreate())
	this.addStatement(SQKeyWrite, this.sqInsert())
	this.addStatement(SQKeyRead, this.sqSelect())

	// If we have primary keys, other operations are possible
	if len(this.PrimaryColumnNames()) > 0 {
		this.addStatement(SQKeyDelete, this.sqDelete())
		this.addStatement(SQKeyGetRowId, this.sqGetRowId())
	}

	// Create index statements
	for _, index := range this.indexes {
		this.addStatement(SQKeyCreate, this.sqIndex(index))
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

func (this *sqclass) Type() reflect.Type {
	return this.t
}

func (this *sqclass) Proto() interface{} {
	return reflect.New(this.t).Interface()
}

func (this *sqclass) Get(k SQKey) []SQStatement {
	this.RWMutex.RLock()
	defer this.RWMutex.RUnlock()

	if st, exists := this.s[k]; exists {
		return st
	} else {
		return nil
	}
}

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

func (this *sqclass) NewIterator(rs SQRows) SQIterator {
	return NewIterator(this.Proto(), rs)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - STATEMENTS

func (this *sqclass) sqCreate() SQTable {
	st := this.CreateTable(this.Columns()...).IfNotExists()
	for _, column := range this.columns {
		if column.Unique {
			st = st.WithUnique(column.Field.Name)
		} else if column.Index {
			st = st.WithIndex(column.Field.Name)
		}
	}
	return st
}

func (this *sqclass) sqIndex(index *sqindex) SQStatement {
	st := N(this.Name()+"_"+index.name).
		WithSchema(this.Schema()).
		CreateIndex(this.Name(), index.cols...).IfNotExists()
	if index.unique {
		st = st.WithUnique()
	}
	return st
}

func (this *sqclass) sqInsert() SQStatement {
	st := this.Insert(this.ColumnNames()...)

	// Add conflict for any primary key field
	st = st.WithConflictUpdate(this.PrimaryColumnNames()...)

	// Add conflict for any unique index
	for _, index := range this.indexes {
		if index.unique {
			st = st.WithConflictUpdate(index.cols...)
		}
	}

	// Return success
	return st
}

func (this *sqclass) sqDelete() SQStatement {
	expr := []interface{}{}
	for _, name := range this.PrimaryColumnNames() {
		expr = append(expr, Q(N(name), "=", P))
	}
	return this.Delete(expr...)
}

func (this *sqclass) sqGetRowId() SQStatement {
	expr := []interface{}{}
	for _, name := range this.PrimaryColumnNames() {
		expr = append(expr, Q(N(name), "=", P))
	}
	return S(this.SQSource).To(N("rowid")).Where(expr...)
}

func (this *sqclass) sqSelect() SQStatement {
	return S(this.SQSource).To(this.ColumnSources()...)
}

func (this *sqclass) addStatement(key SQKey, st SQStatement) {
	this.s[key] = append(this.s[key], st)
}

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

// return bound parameters
func (this *sqclass) params(v interface{}) ([]interface{}, error) {
	// TODO: Needs reworking
	return InsertParams(v)
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
