/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqobj

import (
	"reflect"
	// Frameworks
	"fmt"
	"strconv"
	"strings"

	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	// The sqlite.Connection object
	Conn sq.Connection

	// Language builder
	Lang sq.Language

	// When true, create tables and views if they don't exist when register called
	Create bool
}

type sqobj struct {
	create bool
	conn   sq.Connection
	lang   sq.Language
	log    gopi.Logger
	class  map[string]map[string]*sqclass
}

type sqclass struct {
	name, pkgpath string
	object        bool
	columns       []sq.Column
	conn          sq.Connection
	log           gopi.Logger

	// Statements
	insert  sq.InsertOrReplace
	replace sq.InsertOrReplace
	update  sq.Statement
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	DEFAULT_STRUCT_TAG = "sql"
	SQLITE_PKGPATH     = "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Config) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<sqobj.Open>{ config=%+v }", config)

	this := new(sqobj)
	this.log = logger
	this.create = config.Create
	if conn := config.Conn; conn == nil {
		return nil, gopi.ErrBadParameter
	} else {
		this.conn = conn
	}
	if lang := config.Lang; lang == nil {
		return nil, gopi.ErrBadParameter
	} else {
		this.lang = config.Lang
	}

	// Make a map of classes
	this.class = make(map[string]map[string]*sqclass)

	// Success
	return this, nil
}

func (this *sqobj) Close() error {
	this.log.Debug("<sqobj.Close>{ conn=%v }", this.conn)

	// Release resources
	this.conn = nil
	this.lang = nil
	this.class = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqobj) String() string {
	classes := []string{}
	for k := range this.class {
		for _, sqclass := range this.class[k] {
			classes = append(classes, strconv.Quote(sqclass.Name()))
		}
	}
	return fmt.Sprintf("<sqobj>{ classes=[ %v ] }", strings.Join(classes, ","))
}

////////////////////////////////////////////////////////////////////////////////
// REGISTER

func (this *sqobj) RegisterStruct(v interface{}) (sq.StructClass, error) {
	this.log.Debug2("<sqobj.RegisterStruct>{ %T }", v)

	var class *sqclass
	if columns, err := this.ReflectStruct(v); err != nil {
		return nil, err
	} else if len(columns) == 0 {
		this.log.Warn("Struct without columns is unsupported")
		return nil, sq.ErrUnsupportedType
	} else if name, pkgpath := this.reflectName(v); name == "" {
		this.log.Warn("Struct without name is unsupported")
		return nil, sq.ErrUnsupportedType
	} else if class = this.registeredClass(name, pkgpath); class != nil {
		this.log.Warn("Duplicate registration for %v/%v", pkgpath, name)
		return nil, gopi.ErrBadParameter
	} else if class = this.NewClass(name, pkgpath, this.reflectStructObjectField(v, "RowId") != nil, columns); class == nil {
		return nil, gopi.ErrBadParameter
	} else if this.isExistingTable(class.Name()) == false {
		if this.create == false {
			return nil, sq.ErrNotFound
		} else if st := this.lang.NewCreateTable(class.Name(), columns...); st == nil {
			return nil, gopi.ErrBadParameter
		} else if _, err := this.conn.Do(st.IfNotExists()); err != nil {
			return nil, fmt.Errorf("%v (table %v)", err, strconv.Quote(class.Name()))
		}
	}

	// Register the class here
	if _, exists := this.class[class.pkgpath]; exists == false {
		this.class[class.pkgpath] = make(map[string]*sqclass)
	}
	this.class[class.pkgpath][class.name] = class

	// Return the class
	return class, nil
}

////////////////////////////////////////////////////////////////////////////////
// INSERT, REPLACE, UPDATE

// Insert or replace structs, rollback on error
func (this *sqobj) Write(flags sq.Flag, v ...interface{}) (uint64, error) {
	this.log.Debug2("<sqobj.Write>{ flags=%v num_objects=%v }", flags, len(v))
	// Check for v
	if len(v) == 0 {
		return 0, gopi.ErrBadParameter
	}

	// Map objects to classes
	classmap := make([]*sqclass, len(v))
	for i, value := range v {
		if name, pkgpath := this.reflectName(value); name == "" {
			this.log.Warn("Insert: No struct name")
			return 0, gopi.ErrBadParameter
		} else if class_ := this.registeredClass(name, pkgpath); class_ == nil {
			this.log.Warn("Insert: No registered class for %v (in path %v)", name, strconv.Quote(pkgpath))
			return 0, gopi.ErrBadParameter
		} else {
			classmap[i] = class_
		}
	}

	// Perform the action in a transaction, each object's
	affected_rows := uint64(0)
	if err := this.conn.Txn(func(txn sq.Transaction) error {
		for i, v_ := range v {
			if class := classmap[i]; class == nil {
				return gopi.ErrAppError
			} else if args := class.BoundArgs(v_); args == nil {
				return gopi.ErrAppError
			} else if st, args := class.statement(flags, v_); st == nil || args == nil {
				return sq.ErrUnsupportedType
			} else if r, err := txn.Do(st, args...); err != nil {
				return err
			} else if class.object && reflect.ValueOf(v_).Kind() == reflect.Ptr {
				if field := this.reflectStructObjectField(v_, "RowId"); field != nil {
					field.SetInt(r.LastInsertId)
				}
				affected_rows += r.RowsAffected
			} else {
				affected_rows += r.RowsAffected
			}
		}
		// Success
		return nil
	}); err != nil {
		return affected_rows, err
	} else {
		return affected_rows, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// DELETE

func (this *sqobj) Delete(v ...interface{}) (uint64, error) {
	this.log.Debug2("<sqobj.Delete>{ num_objects=%v }", len(v))

	// Check for v
	if len(v) == 0 {
		return 0, gopi.ErrBadParameter
	}

	return 0, gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *sqobj) isExistingTable(name string) bool {
	for _, table := range this.conn.Tables() {
		if table == name {
			return true
		}
	}
	return false
}

func (this *sqobj) registeredClass(name, pkgpath string) *sqclass {
	if classes, exists := this.class[pkgpath]; exists == false {
		return nil
	} else if class, exists := classes[name]; exists == false {
		return nil
	} else {
		return class
	}
}
