/*
	SQLite client
	(c) Copyright David Thorpe 2017
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package v3

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	sqlite "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACE

// Reflect returns a description of a struct as a set of columns
func (this *client) Reflect(v interface{}) ([]sqlite.Column, error) {
	// Dereference the pointer
	v2 := reflect.ValueOf(v)
	for v2.Kind() == reflect.Ptr {
		v2 = v2.Elem()
	}
	// If not a stuct then return
	if v2.Kind() != reflect.Struct {
		return nil, errors.New("Called Reflect on a non-struct")
	}
	// Enumerate struct fields
	columns := make([]sqlite.Column, v2.NumField())
	for i := 0; i < v2.Type().NumField(); i++ {
		if column, err := columnFor(v2, i); err != nil {
			return nil, err
		} else {
			columns[i] = column
		}
	}
	// Return columns
	return columns, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func columnFor(structValue reflect.Value, i int) (*column, error) {
	// Create the column
	col := &column{
		n:  nameFor(structValue.Type().Field(i)),
		t:  typeFor(structValue.Field(i)),
		f:  sqlite.FLAG_NONE,
		kv: tagsFor(structValue.Type().Field(i)),
	}

	// Iterate through the tag key/value pairs
	for k, v := range col.kv {
		if flag, err := flagsFor(k, v); err != nil {
			return nil, err
		} else {
			col.f |= flag
		}
	}

	return col, nil
}

func nameFor(field reflect.StructField) string {
	return field.Name
}

func typeFor(v reflect.Value) sqlite.Type {
	switch v.Kind() {
	case reflect.String:
		return sqlite.TYPE_TEXT
	case reflect.Bool:
		return sqlite.TYPE_BOOL
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return sqlite.TYPE_UINT
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return sqlite.TYPE_INT
	case reflect.Float32, reflect.Float64:
		return sqlite.TYPE_FLOAT
	case reflect.Struct:
		if _, ok := v.Interface().(time.Time); ok {
			return sqlite.TYPE_TIME
		} else {
			return sqlite.TYPE_NONE
		}
	default:
		if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
			return sqlite.TYPE_NONE
		} else if _, ok := v.Interface().([]byte); ok {
			return sqlite.TYPE_BLOB
		} else {
			return sqlite.TYPE_NONE
		}
	}
}

func tagsFor(field reflect.StructField) map[string]string {
	if tag, ok := field.Tag.Lookup(_V3_TAG); ok == false {
		// No tag
		return map[string]string{}
	} else if fields := strings.Split(tag, ";"); len(fields) != 0 {
		// Probably has tags
		tags := make(map[string]string, len(fields))
		for _, field := range fields {
			if kv := strings.SplitN(field, ":", 2); len(kv) > 0 && kv[0] != "" {
				key := strings.TrimSpace(kv[0])
				if len(kv) == 1 {
					tags[key] = ""
				} else {
					tags[key] = strings.TrimSpace(kv[1])
				}
			}
		}
		return tags
	} else {
		// No tag
		return map[string]string{}
	}
}

// flagsFor returns any flag modifiers, will return FLAG_NONE if
// a flag does not modify the flags
func flagsFor(key string, value string) (sqlite.Flag, error) {
	switch strings.ToLower(key) {
	case "not null", "notnull", "not_null":
		if value != "" {
			return sqlite.FLAG_NONE, fmt.Errorf("Tag '%v' cannot have a value", key)
		}
		return sqlite.FLAG_NOT_NULL, nil
	case "name", "type":
		return sqlite.FLAG_NONE, nil
	case "primary_key", "primary key", "primary":
		return sqlite.FLAG_PRIMARY_KEY, nil
	default:
		return sqlite.FLAG_NONE, fmt.Errorf("Unknown Tag: '%v'", key)
	}
}

/*
func parseFlag(flag string) (sqlite.Flag, error) {
	switch flag {
	case
		return sqlite.FLAG_NOT_NULL, nil
	case "unique_key", "unique key", "unique":
		return sqlite.FLAG_UNIQUE, nil
	case "primary_key", "primary key", "primarykey", "primary":
		return sqlite.FLAG_PRIMARY_KEY, nil
	default:
		return sqlite.FLAG_NONE, fmt.Errorf("Unknown flag '%v'", flag)
	}
}
*/
