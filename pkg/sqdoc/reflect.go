package sqlite

import (
	"reflect"
	"strings"
	"time"
	"unicode"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ReflectStruct returns sql columns for a struct or pointer to a struct
// or returns nil if the structure is not supported
func ReflectStruct(v interface{}, tag string) []sqlite.SQColumn {
	// Dereference the pointer
	v_ := reflect.ValueOf(v)
	for v_.Kind() == reflect.Ptr {
		v_ = v_.Elem()
	}
	// If not a stuct then return
	if v_.Kind() != reflect.Struct {
		return nil
	}
	// Enumerate struct fields
	columns := make([]sqlite.SQColumn, 0, v_.Type().NumField())
	for i := 0; i < v_.Type().NumField(); i++ {
		if column := reflectField(v_.Type().Field(i), tag); column != nil {
			columns = append(columns, column)
		}
	}

	// Return success
	return columns
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func reflectField(field reflect.StructField, name string) nginx.SQColumn {
	var col column

	// Private or anonymous fields not supported
	if field.Anonymous || unicode.IsLower(rune(field.Name[0])) {
		return nil
	}

	// Set the field name
	tags := strings.Split(field.Tag.Get(name), ",")
	if tags[0] == "-" {
		return nil
	} else if tags[0] == "" {
		col.name = field.Name
	} else {
		col.name = tags[0]
	}
	// Set column fields from tags
	for _, tag := range tags[1:] {
		tag = strings.ToUpper(tag)
		switch {
		case isReservedType(tag):
			if col.decltype == "" {
				col.decltype = tag
			}
		case tag == "PRIMARY":
			col.primary = true
		case tag == "NULL":
			col.nullable = true
		case tag == "NOT NULL":
			col.nullable = false
		}
	}
	// Set column decltype from field type
	if col.decltype == "" {
		switch field.Type.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			col.decltype = "INTEGER"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			col.decltype = "INTEGER"
		case reflect.Float32, reflect.Float64:
			col.decltype = "FLOAT"
		case reflect.Bool:
			col.decltype = "INTEGER"
		case reflect.String:
			col.decltype = "TEXT"
		case reflect.Slice:
			if field.Type == reflect.TypeOf([]byte{}) {
				col.decltype = "BLOB"
				col.nullable = true
			}
		case reflect.Struct:
			if field.Type == reflect.TypeOf(time.Time{}) {
				col.decltype = "DATETIME"
			}
		}
	}

	// If decltype is still empty then return nil
	if col.decltype == "" {
		return nil
	}

	// Return success
	return &col
}
