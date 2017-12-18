/*
	SQLite client
	(c) Copyright David Thorpe 2017
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package v3

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	gopi "github.com/djthorpe/gopi"
	sqlite "github.com/djthorpe/sqlite"
	sqlite_driver "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Client defines the configuration parameters for connecting to SQLite Database
type Client struct {
	DSN string
}

type client struct {
	log  gopi.Logger
	conn driver.Conn
}

type column struct {
	n string
	t sqlite.Type
	f sqlite.Flag
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open returns a client object
func (config Client) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<sqlite.v3.Client>Open{ dsn='%v' }", config.DSN)

	this := new(client)
	this.log = log

	d := &sqlite_driver.SQLiteDriver{}
	if _, err := d.Open(config.DSN); err != nil {
		return nil, err
	}

	// Return success
	return this, nil
}

// Close releases any resources associated with the client connection
func (this *client) Close() error {
	this.log.Debug("<sqlite.v3.Client>Close{ }")
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// INTERFACE

// Return description of rows
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
	/*
		if fieldTag, ok := fieldType.Tag.Lookup("sql"); ok {
			if tagName, tagType, tagFlags, err := parseTag(fieldTag); err != nil {
				return nil, err
			} else {
				// Override column information from tag
				if tagName != "" {
					column.n = tagName
				}
				if tagType != sqlite.TYPE_NONE {
					column.t = tagType
				}
				column.f |= tagFlags
			}
		}*/

	fmt.Println(columns)
	return columns, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func columnFor(structValue reflect.Value, i int) (*column, error) {
	col := &column{
		n: nameFor(structValue.Type().Field(i)),
		t: typeFor(structValue.Field(i)),
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
		} else if _, ok := v.Interface().(time.Duration); ok {
			return sqlite.TYPE_DURATION
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

func parseTag(tag string) (string, sqlite.Type, sqlite.Flag, error) {
	f := sqlite.FLAG_NONE
	n := ""
	if fields := strings.Split(tag, ";"); len(fields) != 0 {
		for _, field := range fields {
			if kv := strings.SplitN(field, ":", 2); len(kv) > 0 && kv[0] != "" {
				switch kv[0] {
				case "name":
					if len(kv) != 2 {
						return "", 0, 0, fmt.Errorf("Parse error in tag '%v': Invalid name", tag)
					} else {
						n = kv[1]
					}
					break
				case "type":
					break
				default:
					if len(kv) != 1 {
						return "", 0, 0, fmt.Errorf("Parse error in tag '%v': Invalid flag '%v'", tag, kv[0])
					} else if flag, err := parseFlag(kv[0]); err != nil {
						return "", 0, 0, fmt.Errorf("Parse error in tag '%v': %v", tag, err)
					} else {
						f |= flag
					}
				}
				fmt.Printf("KV %v %v\n", len(kv), kv)
			}
		}
	}
	return n, 0, f, nil
}

func parseFlag(flag string) (sqlite.Flag, error) {
	switch flag {
	case "not null", "notnull", "not_null":
		return sqlite.FLAG_NOT_NULL, nil
	case "unique_key", "unique key", "unique":
		return sqlite.FLAG_UNIQUE, nil
	case "primary_key", "primary key", "primarykey", "primary":
		return sqlite.FLAG_PRIMARY_KEY, nil
	default:
		return sqlite.FLAG_NONE, fmt.Errorf("Unknown flag '%v'", flag)
	}
}

////////////////////////////////////////////////////////////////////////////////
// COLUMN

func (this *column) Name() string {
	return this.n
}

func (this *column) Type() sqlite.Type {
	return this.t
}

func (this *column) Flags() sqlite.Flag {
	return this.f
}

func (this *column) String() string {
	return fmt.Sprintf("<sqlite.v3.Column>{ name=%v type=%v flags=%v }", this.n, this.t, this.f)
}
