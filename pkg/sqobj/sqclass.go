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
	s map[SQKey][]SQStatement

	// All columns, primary key, unique and index column names
	col     []SQColumn
	primary []*marshaler.Field
	unique  []string
	index   []string
}

type sqindex struct {
	name   string
	unique bool
	cols   []string
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewClass(name, schema string, proto interface{}) *sqclass {
	this := new(sqclass)
	this.s = make(map[SQKey][]SQStatement)

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

	// Get columns for the prototype
	fields := marshaler.NewEncoder(TagName).Reflect(proto)
	if len(fields) == 0 {
		return nil
	}
	for _, field := range fields {
		c := C(field.Name).WithType(decltype(field.Type))
		for _, tag := range field.Tags {
			if IsSupportedType(tag) {
				c = c.WithType(strings.ToUpper(tag))
			} else if isNotNull(tag) {
				c = c.NotNull()
			} else if isPrimary(tag) {
				c = c.NotNull()
				c = c.WithPrimary()
				this.primary = append(this.primary, field)
			} else if isAutoincrement(tag) {
				c = c.NotNull()
				c = c.WithPrimary()
				c = c.WithAutoIncrement()
				this.primary = append(this.primary, field)
			} else if isUnique(tag) {
				this.unique = append(this.unique, field.Name)
			} else if isIndex(tag) {
				this.index = append(this.index, field.Name)
			}
		}
		this.col = append(this.col, c)
	}

	// Create statments for create,insert and delete
	this.addStatement(SQKeyCreate, this.sqCreate())
	this.addStatement(SQKeyWrite, this.sqInsert())

	// If we have primary keys, other operations are possible
	if len(this.primary) > 0 {
		this.addStatement(SQKeyDelete, this.sqDelete())
		this.addStatement(SQKeyGetRowId, this.sqGetRowId())
	}

	// Get indexes for the prototype
	indexes := make(map[string]*sqindex)
	for _, field := range fields {
		for _, tag := range field.Tags {
			if strings.HasPrefix(tag, "index:") {
				if _, exists := indexes[tag]; exists {
					indexes[tag].cols = append(indexes[tag].cols, field.Name)
				} else {
					indexes[tag] = &sqindex{tag[6:], false, []string{field.Name}}
				}
			} else if strings.HasPrefix(tag, "unique:") {
				if _, exists := indexes[tag]; exists {
					indexes[tag].cols = append(indexes[tag].cols, field.Name)
				} else {
					indexes[tag] = &sqindex{tag[7:], true, []string{field.Name}}
				}
			}
		}
	}

	// Create index statements
	for _, index := range indexes {
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
	str += fmt.Sprintf(" cols=%q", this.col)
	str += fmt.Sprintf(" primary=%q", namesForFields(this.primary))
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

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - STATEMENTS

func (this *sqclass) sqCreate() SQStatement {
	st := this.CreateTable(this.col...).IfNotExists()
	for _, index := range this.index {
		st = st.WithIndex(index)
	}
	for _, unique := range this.unique {
		st = st.WithUnique(unique)
	}
	return st
}

func (this *sqclass) sqIndex(index *sqindex) SQStatement {
	st := N(this.Name()+"_"+index.name).
		CreateIndex(this.Name(), index.cols...).IfNotExists()
	if index.unique {
		st = st.WithUnique()
	}
	return st
}

func (this *sqclass) sqInsert() SQStatement {
	st := this.Insert(namesForColumns(this.col)...)
	if len(this.primary) > 0 {
		st = st.WithConflictUpdate(namesForFields(this.primary)...)
	}
	return st
}

func (this *sqclass) sqDelete() SQStatement {
	expr := []interface{}{}
	for _, name := range namesForFields(this.primary) {
		expr = append(expr, Q(N(name), "=", P))
	}
	return this.Delete(expr...)
}

func (this *sqclass) sqGetRowId() SQStatement {
	expr := []interface{}{}
	for _, name := range namesForFields(this.primary) {
		expr = append(expr, Q(N(name), "=", P))
	}
	return S(this.SQSource).To(N("rowid")).Where(expr...)
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
	for _, field := range this.primary {
		if v, err_ := sqlite.BoundValue(r.Field(field.Index)); err != nil {
			err = multierror.Append(err, err_)
		} else {
			result = append(result, v)
		}
	}
	// Return any errors
	return result, err
}
