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
	t2 := v2.Type()
	for i := 0; i < t2.NumField(); i++ {
		fieldType := t2.Field(i)
		column, err := columnFor(fieldType)
		if err != nil {
			return nil, err
		}
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
		}
		columns[i] = column
	}
	return columns, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func columnFor(f reflect.StructField) (*column, error) {
	fmt.Printf("name %v => %v type %v\n", f.Name, f.PkgPath, typeFor(f.Type))
	return &column{
		n: f.Name,
		t: typeFor(f.Type),
	}, nil
}

func typeFor(t reflect.Type) sqlite.Type {
	switch {
	case t.Kind() == reflect.String:
		return sqlite.TYPE_TEXT
	case t.Kind() == reflect.Bool:
		return sqlite.TYPE_BOOL
	case t.Kind() == reflect.Uint || t.Kind() == reflect.Uint8 || t.Kind() == reflect.Uint16 || t.Kind() == reflect.Uint32 || t.Kind() == reflect.Uint64:
		return sqlite.TYPE_UINT
	case t.Kind() == reflect.Int || t.Kind() == reflect.Int8 || t.Kind() == reflect.Int16 || t.Kind() == reflect.Int32 || t.Kind() == reflect.Int64:
		return sqlite.TYPE_INT
	case t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64:
		return sqlite.TYPE_FLOAT
	default:
		return sqlite.TYPE_NONE
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
