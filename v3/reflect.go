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
	columns := make([]sqlite.Column, 0, v2.NumField())
	for i := 0; i < v2.Type().NumField(); i++ {
		if column, err := columnFor(v2, i); err != nil {
			return nil, err
		} else if column != nil {
			columns = append(columns, column)
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
		n: nameFor(structValue.Type().Field(i)),
		t: typeFor(structValue.Field(i)),
		f: flagsFor(structValue.Type().Field(i)),
	}

	// If f is nil, then we return nil so we ignore this column,
	// the tag is set as sql:"-" which means to ignore the field
	if col.f == nil {
		return nil, nil
	}

	// If f contains FLAG_NONE then there has been an error
	// interpreting the flags
	if _, exists := col.f[sqlite.FLAG_NONE]; exists {
		return nil, fmt.Errorf("%v: invalid tag fields", col.n)
	}

	// Return column
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

func flagsFor(field reflect.StructField) map[sqlite.Flag]string {
	if tag, ok := field.Tag.Lookup(_V3_TAG); ok == false {
		// No tag
		return map[sqlite.Flag]string{}
	} else if tag == "-" {
		// Ignore - return nil
		return nil
	} else if fields := strings.Split(tag, ";"); len(fields) != 0 {
		// Probably has tags
		tags := make(map[sqlite.Flag]string, len(fields))
		for _, field := range fields {
			if kv := strings.SplitN(field, ":", 2); len(kv) > 0 && kv[0] != "" {
				flag := flagFor(strings.TrimSpace(kv[0]))
				if len(kv) == 1 {
					tags[flag] = ""
				} else {
					tags[flag] = strings.TrimSpace(kv[1])
				}
			}
		}
		return tags
	} else {
		// No tag
		return map[sqlite.Flag]string{}
	}
}

// flagsFor returns flag for key or FLAG_NONE if the key was invalud
func flagFor(key string) sqlite.Flag {
	switch strings.ToLower(key) {
	case "not null", "notnull", "not_null":
		return sqlite.FLAG_NOT_NULL
	case "name":
		return sqlite.FLAG_NAME
	case "type":
		return sqlite.FLAG_TYPE
	case "primary_key", "primary key", "primary":
		return sqlite.FLAG_PRIMARY_KEY
	case "key", "index":
		return sqlite.FLAG_INDEX_KEY
	case "unique", "unique key", "unique_key":
		return sqlite.FLAG_UNIQUE_KEY
	default:
		return sqlite.FLAG_NONE
	}
}
