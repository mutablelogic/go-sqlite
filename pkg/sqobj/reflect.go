package sqobj

import (
	"reflect"
	"strings"
	"time"

	// Modules
	marshaler "github.com/djthorpe/go-marshaler"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
	multierror "github.com/hashicorp/go-multierror"

	// Import namespaces
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type index struct {
	name   string
	unique bool
	cols   []string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	timeType = reflect.TypeOf(time.Time{})
	blobType = reflect.TypeOf([]byte{})
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// CreateTable returns a CREATE TABLE statement for the given struct
// or nil if the argument is not a pointer to a struct or has no fields which are exported
func CreateTable(source SQSource, v interface{}) SQTable {
	c, i, u := structCols(v)
	if len(c) == 0 {
		return nil
	}
	s := source.CreateTable(c...)
	for _, index := range i {
		s = s.WithIndex(index)
	}
	for _, unique := range u {
		s = s.WithUnique(unique)
	}
	return s
}

// CreateIndexes returns CREATE INDEX statements for the given struct
// or nil if the argument is not a pointer to a struct or has no fields which are exported
func CreateIndexes(source SQSource, v interface{}) []SQIndexView {
	var result []SQIndexView
	for _, index := range structIndexes(v) {
		index_source := source.WithName(source.Name() + "_" + index.name)
		q := index_source.CreateIndex(source.Name(), index.cols...)
		if index.unique {
			q = q.WithUnique()
		}
		result = append(result, q)
	}
	return result
}

// CreateTableAndIndexes returns statements for creating table and indexes
// and returns them, with the CreateTable being first
func CreateTableAndIndexes(source SQSource, ifnotexists bool, v interface{}) []SQStatement {
	result := []SQStatement{}

	// Create table
	t := CreateTable(source, v)
	if ifnotexists {
		t = t.IfNotExists()
	}
	result = append(result, t)

	// Create indexes
	for _, index := range CreateIndexes(source, v) {
		if ifnotexists {
			index = index.IfNotExists()
		}
		result = append(result, index)
	}

	// Return statements
	return result
}

// InsertRow returns an INSERT statement for the given struct or nil if the
// argument is not a pointer to a struct or has no fields which are exported
func InsertRow(name string, v interface{}) SQInsert {
	c, _, _ := structCols(v)
	if c == nil || len(c) == 0 {
		return nil
	}
	return N(name).Insert(namesForColumns(c)...)
}

// ReplaceRow returns an INSERT OR REPLACE statement for the given struct or nil if the
// argument is not a pointer to a struct or has no fields which are exported
func ReplaceRow(name string, v interface{}) SQInsert {
	c, _, _ := structCols(v)
	if c == nil || len(c) == 0 {
		return nil
	}
	return N(name).Replace(namesForColumns(c)...)
}

// InsertParams returns the parameters from a struct to use for an insert statement or
// returns an error
func InsertParams(v interface{}) ([]interface{}, error) {
	fields := marshaler.NewEncoder(TagName).Reflect(v)
	if fields == nil {
		return nil, ErrBadParameter
	}
	var err error
	result := make([]interface{}, len(fields))
	for i, field := range fields {
		if v, err_ := sqlite.BoundValue(field.Value); err_ != nil {
			err = multierror.Append(err, err_)
		} else {
			result[i] = v
		}
	}
	return result, err
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func structCols(v interface{}) ([]SQColumn, []string, []string) {
	fields := marshaler.NewEncoder(TagName).Reflect(v)
	if fields == nil {
		return nil, nil, nil
	}
	result := make([]SQColumn, 0, len(fields))
	indexes := make([]string, 0, len(fields))
	uniques := make([]string, 0, len(fields))
	for _, field := range fields {
		c := C(field.Name).WithType(decltype(field.Type))
		for _, tag := range field.Tags {
			if IsSupportedType(tag) {
				c = c.WithType(strings.ToUpper(tag))
			} else if isNotNull(tag) {
				c = c.NotNull()
			} else if isPrimary(tag) {
				c = c.WithPrimary()
			} else if isUnique(tag) {
				uniques = append(uniques, field.Name)
			} else if isIndex(tag) {
				indexes = append(indexes, field.Name)
			} else if isAutoincrement(tag) {
				c = c.WithAutoIncrement()
			}
		}
		result = append(result, c)
	}

	return result, indexes, uniques
}

func structIndexes(v interface{}) map[string]*index {
	result := map[string]*index{}
	fields := marshaler.NewEncoder(TagName).Reflect(v)
	if fields == nil {
		return nil
	}
	for _, field := range fields {
		for _, tag := range field.Tags {
			if strings.HasPrefix(tag, "index:") {
				if _, exists := result[tag]; exists {
					result[tag].cols = append(result[tag].cols, field.Name)
				} else {
					result[tag] = &index{tag[6:], false, []string{field.Name}}
				}
			} else if strings.HasPrefix(tag, "unique:") {
				if _, exists := result[tag]; exists {
					result[tag].cols = append(result[tag].cols, field.Name)
				} else {
					result[tag] = &index{tag[7:], true, []string{field.Name}}
				}
			}
		}
	}
	return result
}

func isNotNull(tag string) bool {
	tag = strings.TrimSpace(strings.ToUpper(tag))
	return tag == "NOT NULL" || tag == "NOTNULL"
}

func isPrimary(tag string) bool {
	tag = strings.TrimSpace(strings.ToUpper(tag))
	return tag == "PRI" || tag == "PRIMARY" || tag == "PRIMARY KEY"
}

func isAutoincrement(tag string) bool {
	tag = strings.TrimSpace(strings.ToUpper(tag))
	return tag == "AUTOINCREMENT"
}

func isUnique(tag string) bool {
	tag = strings.TrimSpace(strings.ToUpper(tag))
	return tag == "UNIQUE"
}

func isForeign(tag string) bool {
	tag = strings.TrimSpace(strings.ToUpper(tag))
	return tag == "FOREIGN"
}

func isIndex(tag string) bool {
	tag = strings.TrimSpace(strings.ToUpper(tag))
	return tag == "INDEX"
}

func decltype(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "FLOAT"
	case reflect.Bool:
		return "INTEGER"
	default:
		if t == timeType {
			return "TIMESTAMP"
		}
		if t == blobType {
			return "BLOB"
		}
		return "TEXT"
	}
}

func namesForColumns(cols []SQColumn) []string {
	result := make([]string, 0, len(cols))
	for _, col := range cols {
		if name := col.Name(); name != "" {
			result = append(result, name)
		}
	}
	return result
}

func sourcesForColumns(cols []SQColumn) []SQSource {
	result := make([]SQSource, 0, len(cols))
	for _, col := range cols {
		if name := col.Name(); name != "" {
			result = append(result, N(name))
		}
	}
	return result
}

func primaryForColumns(cols []SQColumn) []string {
	var result []string
	for _, col := range cols {
		if col.Primary() != "" {
			result = append(result, col.Name())
		}
	}
	return result
}

func namesForFields(cols []*marshaler.Field) []string {
	result := map[string]bool{}
	for _, col := range cols {
		result[col.Name] = true
	}
	return toArrayKeys(result)
}

func arrayContains(v []string, elem string) bool {
	for _, v := range v {
		if v == elem {
			return true
		}
	}
	return false
}

func toArrayKeys(v map[string]bool) []string {
	result := make([]string, 0, len(v))
	for k, _ := range v {
		result = append(result, k)
	}
	return result
}

// valueOf returns a struct value or nil if not valid
func valueOf(v interface{}) reflect.Value {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		return rv
	} else {
		return reflect.ValueOf(nil)
	}
}
