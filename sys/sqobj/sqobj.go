/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqobj

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	// Frameworks
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
	tablename     string
	object        bool
	columns       []sq.Column
	keys          []string
	conn          sq.Connection
	lang          sq.Language
	log           gopi.Logger

	// Statements
	insert  sq.InsertOrReplace
	replace sq.InsertOrReplace
	update  sq.Update
	delete  sq.Delete
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
	return fmt.Sprintf("<sqobj>{ classes=[ %v ] conn=%v }", strings.Join(classes, ","), this.conn)
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *sqobj) Conn() sq.Connection {
	return this.conn
}

func (this *sqobj) Lang() sq.Language {
	return this.lang
}

////////////////////////////////////////////////////////////////////////////////
// REGISTER

func (this *sqobj) RegisterStruct(v interface{}) (sq.StructClass, error) {
	this.log.Debug2("<sqobj.RegisterStruct>{ %T }", v)

	var class *sqclass
	if columns, err := this.reflectStruct(v); err != nil {
		return nil, err
	} else if len(columns) == 0 {
		return nil, fmt.Errorf("%w: Struct without columns is unsupported", sq.ErrUnsupportedType)
	} else if name, pkgpath := this.reflectName(v); name == "" {
		return nil, fmt.Errorf("%w: Anonymous Structs are unsupported", sq.ErrUnsupportedType)
	} else if class = this.registeredClass(name, pkgpath); class != nil {
		return nil, fmt.Errorf("%w: Duplicate registration for %v/%v", gopi.ErrBadParameter, pkgpath, name)
	} else if tablename, err := this.reflectTableName(v); err != nil {
		return nil, fmt.Errorf("%w: Error for TableName method", err)
	} else if class = this.NewClass(name, pkgpath, tablename, reflectStructObjectField(reflect.ValueOf(v)) != nil, columns); class == nil {
		return nil, fmt.Errorf("%w: NewClass failed", gopi.ErrBadParameter)
	} else if this.isExistingTable(class.TableName()) == false {
		if this.create == false {
			return nil, fmt.Errorf("%w: Table %v does not exist", sq.ErrNotFound, strconv.Quote(class.TableName()))
		} else if st := this.lang.NewCreateTable(class.TableName(), columns...); st == nil {
			return nil, gopi.ErrBadParameter
		} else if _, err := this.conn.Do(st.IfNotExists()); err != nil {
			return nil, fmt.Errorf("%w: Table %v execution error", err, strconv.Quote(class.TableName()))
		}
	} else if table_cols, err := this.conn.ColumnsForTable(class.TableName(), ""); err != nil {
		return nil, fmt.Errorf("%w: ColumnsForTable failed", err)
	} else {
		// Make a map of table columns for quick referencing
		table_col_map := make(map[string]sq.Column, len(table_cols))
		for _, table_col := range table_cols {
			table_col_map[table_col.Name()] = table_col
		}
		// Ensure the table includes columns
		for _, col := range class.ColumnNames() {
			if table_col, exists := table_col_map[col]; exists == false {
				return nil, fmt.Errorf("%w: Unsupported table %v, column %v does not exist", sq.ErrUnsupportedType, strconv.Quote(class.TableName()), strconv.Quote(col))
			} else if class.DeclTypeForColumn(col) != table_col.DeclType() {
				return nil, fmt.Errorf("%w: Type mismatch for column %v on table %v", sq.ErrUnsupportedType, strconv.Quote(col), strconv.Quote(class.TableName()))
			}
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
// WRITE FUNCTION

func (this *sqobj) Write(flags sq.Flag, v ...interface{}) (uint64, error) {
	this.log.Debug2("<sqobj.Write>{ flags=%v num_objects=%v }", flags, len(v))

	// Check for v
	if len(v) == 0 {
		return 0, fmt.Errorf("%w: No objects to write", gopi.ErrBadParameter)
	}
	// Check operation
	if flags&sq.FLAG_OP_MASK == sq.FLAG_NONE {
		return 0, fmt.Errorf("%w: Invalid flag parameter", gopi.ErrBadParameter)
	}
	// Make an array of classes for the objects
	if classes, err := this.classesFor(v); err != nil {
		return 0, err
	} else {
		// Perform the operation within a transaction
		total_affected_rows := uint64(0)
		if err := this.conn.Txn(func(txn sq.Transaction) error {
			for i, v_ := range v {
				if class := classes[i]; class == nil {
					return gopi.ErrAppError
				} else if r, err := class.op_write(txn, flags, v_); err != nil {
					return err
				} else {
					total_affected_rows += r.RowsAffected
				}
			}
			// Success
			return nil
		}); err != nil {
			return 0, err
		}
		// Return success
		return total_affected_rows, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// DELETE

func (this *sqobj) Delete(v ...interface{}) (uint64, error) {
	this.log.Debug2("<sqobj.Delete>{ num_objects=%v }", len(v))
	// Check for v
	if len(v) == 0 {
		return 0, fmt.Errorf("%w: No objects to delete", gopi.ErrBadParameter)
	}
	// Make an array of classes for the objects
	if classes, err := this.classesFor(v); err != nil {
		return 0, err
	} else {
		// Perform the operation within a transaction
		total_affected_rows := uint64(0)
		if err := this.conn.Txn(func(txn sq.Transaction) error {
			for i, v_ := range v {
				if class := classes[i]; class == nil {
					return gopi.ErrAppError
				} else if r, err := class.op_delete(txn, v_); err != nil {
					return err
				} else {
					total_affected_rows += r.RowsAffected
				}
			}
			// Success
			return nil
		}); err != nil {
			return 0, err
		}
		// Return success
		return total_affected_rows, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// COUNT

func (this *sqobj) Count(class sq.Class) (uint64, error) {
	this.log.Debug2("<sqobj.Count>{ class=%v }", class)
	if class == nil {
		return 0, gopi.ErrBadParameter
	} else if class_, ok := class.(*sqclass); ok == false {
		return 0, gopi.ErrBadParameter
	} else {
		q := class_.query()
		fmt.Println(q.Query())
	}
	// Return nil
	return 0, nil
}
