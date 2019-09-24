/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqobj

import (
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
	name    string
	columns []sq.Column
	insert  sq.InsertOrReplace
	conn    sq.Connection
	log     gopi.Logger
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	DEFAULT_STRUCT_TAG = "sql"
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

	if columns, err := this.ReflectStruct(v); err != nil {
		return nil, err
	} else if len(columns) == 0 {
		this.log.Warn("Struct without columns is unsupported")
		return nil, sq.ErrUnsupportedType
	} else if name, pkgpath := this.reflectName(v); name == "" {
		this.log.Warn("Struct without name is unsupported")
		return nil, sq.ErrUnsupportedType
	} else if class := this.registeredClass(name, pkgpath); class != nil {
		this.log.Warn("Duplicate registration for %v/%v", pkgpath, name)
		return nil, gopi.ErrBadParameter
	} else if class = this.NewClass(name, pkgpath, columns); class == nil {
		return nil, gopi.ErrBadParameter
	} else if this.isExistingTable(class.Name()) == false {
		if this.create == false {
			return nil, sq.ErrNotFound
		} else if st := this.lang.NewCreateTable(class.Name(), columns...); st == nil {
			return nil, gopi.ErrBadParameter
		} else if _, err := this.conn.Do(st.IfNotExists()); err != nil {
			return nil, fmt.Errorf("%v (table %v)", err, strconv.Quote(class.Name()))
		}
		return class, nil
	} else {
		return class, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// INSERT

func (this *sqobj) Insert(...interface{}) ([]int64, error) {
	return nil, gopi.ErrNotImplemented
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
	return nil
}
