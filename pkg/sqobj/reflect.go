package sqobj

import (
	"reflect"
	"strings"
	"time"

	// Modules
	marshaler "github.com/djthorpe/go-marshaler"
	sqlite "github.com/djthorpe/go-sqlite"
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
func CreateTable(name string, v interface{}) sqlite.SQTable {
	if c := structCols(v); c == nil {
		return nil
	} else {
		return N(name).CreateTable(c...)
	}
}

// CreateIndexes returns CREATE INDEX statements for the given struct
// or nil if the argument is not a pointer to a struct or has no fields which are exported
func CreateIndexes(name string, v interface{}) []sqlite.SQIndexView {
	var result []sqlite.SQIndexView
	for _, index := range structIndexes(v) {
		q := N(index.name).CreateIndex(name, index.cols...)
		if index.unique {
			q = q.WithUnique()
		}
		result = append(result, q)
	}
	return result
}

// InsertRow returns an INSERT statement for the given struct
// or nil if the argument is not a pointer to a struct or has no fields which are exported
func InsertRow(name string, v interface{}) sqlite.SQInsert {
	c := structCols(v)
	if c == nil || len(c) == 0 {
		return nil
	}
	return N(name).Insert(namesForColumns(c)...)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func structCols(v interface{}) []sqlite.SQColumn {
	fields := marshaler.NewEncoder(sqlite.TagName).Reflect(v)
	if fields == nil {
		return nil
	}
	result := make([]sqlite.SQColumn, 0, len(fields))
	for _, field := range fields {
		c := C(field.Name).WithType(decltype(field.Type))
		for _, tag := range field.Tags {
			if sqlite.IsSupportedType(tag) {
				c = c.WithType(strings.ToUpper(tag))
			} else if isNotNull(tag) {
				c = c.NotNull()
			} else if isPrimary(tag) {
				c = c.Primary()
			}
		}
		result = append(result, c)
	}

	return result
}

func structIndexes(v interface{}) map[string]*index {
	result := map[string]*index{}
	fields := marshaler.NewEncoder(sqlite.TagName).Reflect(v)
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

func namesForColumns(cols []sqlite.SQColumn) []string {
	result := make([]string, len(cols))
	for i, col := range cols {
		result[i] = col.Name()
	}
	return result
}
