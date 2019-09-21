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

	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	// The sqlite.Connection object
	Conn sq.Connection

	// When true, create tables and views if they don't exist when register called
	Create bool
}

type sqobj struct {
	create bool
	conn   sq.Connection
	log    gopi.Logger
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
	this.conn = config.Conn
	this.create = config.Create

	// Success
	return this, nil
}

func (this *sqobj) Close() error {
	this.log.Debug("<sqobj.Close>{ conn=%v }", this.conn)

	// Release resources
	this.conn = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqobj) String() string {
	return fmt.Sprintf("<sqobj>{ conn=%v }", this.conn)
}

////////////////////////////////////////////////////////////////////////////////
// REGISTER

func (this *sqobj) RegisterStruct(name string, v interface{}) (sq.StructClass, error) {
	this.log.Debug2("<sqobj.RegisterStruct>{ %T }", v)

	if name == "" {
		return nil, gopi.ErrBadParameter
	} else if columns, err := this.ReflectStruct(v); err != nil {
		return nil, err
	} else if len(columns) == 0 {
		this.log.Warn("No colmns for struct: %v", strconv.Quote(name))
		return nil, sq.ErrUnsupportedType
	} else {
		if this.isExistingTable(name) == false {
			if this.create == false {
				return nil, sq.ErrNotFound
			} else if st := this.conn.NewCreateTable(name, columns...); st == nil {
				return nil, gopi.ErrBadParameter
			} else if _, err := this.conn.Do(st.IfNotExists()); err != nil {
				return nil, fmt.Errorf("%v (table %v)", err, strconv.Quote(name))
			} else {
				this.log.Debug(st.Query(this.conn))
			}
		}
		if class := this.NewClass(name, columns); class == nil {
			return nil, gopi.ErrBadParameter
		} else {
			return class, nil
		}
	}
}

func (this *sqobj) isExistingTable(name string) bool {
	for _, table := range this.conn.Tables() {
		if table == name {
			return true
		}
	}
	return false
}
