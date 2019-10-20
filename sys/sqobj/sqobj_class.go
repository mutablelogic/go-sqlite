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

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqobj) NewClass(name, pkgpath, tablename string, object bool, columns []sq.Column) *sqclass {
	// check incoming parameters
	if name == "" || pkgpath == "" || tablename == "" || len(columns) == 0 {
		return nil
	}

	// determine the keys
	keys := make([]string, 0, 1)
	if object {
		keys = append(keys, "_rowid_")
	} else {
		for _, column := range columns {
			if column.PrimaryKey() {
				keys = append(keys, column.Name())
			}
		}
	}

	// Create class object
	class := &sqclass{name, pkgpath, tablename, object, columns, keys, this.conn, this.lang, this.log, nil, nil, nil, nil, nil}

	// Prepare insert, replace and count statements
	if class.insert = this.lang.Insert(tablename, class.ColumnNames()...); class.insert == nil {
		return nil
	} else if class.replace = this.lang.Replace(tablename, class.ColumnNames()...); class.replace == nil {
		return nil
	} else if class.count = this.lang.NewSelect(this.lang.NewSource(tablename), this.lang.CountAll()); class.count == nil {
		return nil
	}
	// Prepare update and delete if there are keys associated
	if object || len(class.keys) > 0 {
		class.update = this.lang.NewUpdate(tablename, class.names_of(object == false)...).Where(class.key_expression())
		class.delete = this.lang.NewDelete(tablename).Where(class.key_expression())
	}
	// Success
	return class
}

func (this *sqclass) key_expression() sq.Expression {
	if this.object {
		// WHERE _row_id=?
		return this.lang.Equals("_rowid_", this.lang.Arg())
	}
	if len(this.keys) > 0 {
		// WHERE <key>=? AND <key>=? ...
		expr := make([]sq.Expression, len(this.keys))
		for i, row := range this.keys {
			expr[i] = this.lang.Equals(row, this.lang.Arg())
		}
		return this.lang.And(expr...)
	}
	return nil
}

func (this *sqclass) Name() string {
	return this.name
}

func (this *sqclass) TableName() string {
	return this.tablename
}

func (this *sqclass) ColumnNames() []string {
	return this.names_of(false)
}

func (this *sqclass) DeclTypeForColumn(name string) string {
	for _, column := range this.columns {
		if column.Name() == name {
			return column.DeclType()
		}
	}
	return ""
}

// Keys returns the names of the key fields for a struct. For object-based
// struct, this is ["_rowid_"] otherwise it's an array of the primary
// key fields
func (this *sqclass) Keys() []string {
	return this.keys
}

// BoundArgs returns the bound arguments for a particular operation. Where
// the operation is INSERT or INSERT|UPDATE it will return all column values
// for UPDATE it will return all the non-key columns and append the key names and
// values. For DELETE it will return only the key values
func (this *sqclass) BoundArgs(flags sq.Flag, v interface{}) []interface{} {
	// Dereference the pointer
	v_ := reflect.ValueOf(v)
	for v_.Kind() == reflect.Ptr {
		v_ = v_.Elem()
	}
	// If not a stuct then return
	if v_.Kind() != reflect.Struct {
		return nil
	}
	// Compose args based on the operation
	switch flags & sq.FLAG_OP_MASK {
	case sq.FLAG_INSERT, sq.FLAG_INSERT | sq.FLAG_UPDATE:
		return this.values_of(v_, false)
	case sq.FLAG_UPDATE:
		if key_values := this.key_values_of(v_); len(key_values) == 0 {
			return nil
		} else {
			return append(this.values_of(v_, this.object == false), key_values...)
		}
	case sq.FLAG_DELETE:
		return this.key_values_of(v_)
	default:
		// Unknown operation
		return nil
	}
}

func (this *sqclass) statement(flags sq.Flag, v interface{}) (sq.Statement, []interface{}, error) {
	switch {
	case flags&sq.FLAG_OP_MASK == sq.FLAG_INSERT:
		return this.insert, this.BoundArgs(sq.FLAG_INSERT, v), nil
	case flags&sq.FLAG_OP_MASK == sq.FLAG_INSERT|sq.FLAG_UPDATE:
		return this.replace, this.BoundArgs(sq.FLAG_INSERT, v), nil
	case flags&sq.FLAG_OP_MASK == sq.FLAG_UPDATE:
		if this.update == nil {
			return nil, nil, fmt.Errorf("%w: UPDATE not supported for class %v", sq.ErrUnsupportedType, strconv.Quote(this.name))
		}
		if this.object && reflectStructObjectField(reflect.ValueOf(v)).Int() == 0 {
			return nil, nil, fmt.Errorf("%w: UPDATE on invalid object for class %v", gopi.ErrOutOfOrder, strconv.Quote(this.name))
		}
		return this.update, this.BoundArgs(sq.FLAG_UPDATE, v), nil
	case flags&sq.FLAG_OP_MASK == sq.FLAG_DELETE:
		if this.delete == nil {
			return nil, nil, fmt.Errorf("%w: DELETE not supported for class %v", sq.ErrUnsupportedType, strconv.Quote(this.name))
		}
		if this.object && reflectStructObjectField(reflect.ValueOf(v)).Int() == 0 {
			return nil, nil, fmt.Errorf("%w: DELETE on invalid object for class %v", gopi.ErrOutOfOrder, strconv.Quote(this.name))
		}
		return this.delete, this.BoundArgs(sq.FLAG_DELETE, v), nil
	default:
		return nil, nil, fmt.Errorf("%w: Operation not supported for class %v", sq.ErrUnsupportedType, strconv.Quote(this.name))
	}
}

// write performs an insert, replace or update operation and returns the
// Result structure or error
func (this *sqclass) op_write(txn sq.Transaction, flags sq.Flag, v interface{}) (sq.Result, error) {
	if statement, args, err := this.statement(flags, v); err != nil {
		return sq.Result{}, err
	} else if r, err := txn.Do(statement, args...); err != nil {
		return sq.Result{}, err
	} else {
		// Update RowId on insert
		v_ := reflect.ValueOf(v)
		if this.object && v_.Kind() == reflect.Ptr && (flags&sq.FLAG_INSERT) == sq.FLAG_INSERT {
			if field := reflectStructObjectField(v_); field != nil {
				field.SetInt(r.LastInsertId)
			}
		}
		return r, nil
	}
}

// delete an object within a transaxtion, returns the Result structure or error
func (this *sqclass) op_delete(txn sq.Transaction, v interface{}) (sq.Result, error) {
	if statement, args, err := this.statement(sq.FLAG_DELETE, v); err != nil {
		return sq.Result{}, err
	} else if r, err := txn.Do(statement, args...); err != nil {
		return sq.Result{}, err
	} else {
		// Update RowId on delete
		v_ := reflect.ValueOf(v)
		if this.object && v_.Kind() == reflect.Ptr {
			if field := reflectStructObjectField(v_); field != nil {
				field.SetInt(0)
			}
		}
		return r, nil
	}
}

// count the number of objects
func (this *sqclass) op_count(conn sq.Connection) (sq.Rows, error) {
	return conn.Query(this.count)
}

// read objects
func (this *sqclass) op_read(conn sq.Connection, limit, offset uint) (sq.Rows, error) {
	source := this.lang.NewSource(this.tablename)
	st := this.lang.NewSelect(source, this.query_expressions()...).LimitOffset(limit, offset)
	return conn.Query(st)
}

// key_values_of returns the key values for a struct. For object-based
// struct, this is the value of _rowid_ otherwise it's an array of
// the primary key values
func (this *sqclass) key_values_of(v reflect.Value) []interface{} {
	if v.Kind() != reflect.Struct {
		return nil
	}
	if len(this.keys) == 0 {
		return nil
	}
	if this.object {
		if field := reflectStructObjectField(v); field != nil {
			return []interface{}{field.Int()}
		} else {
			return nil
		}
	}
	args := make([]interface{}, 0, len(this.keys))
	for _, column := range this.columns {
		if column.PrimaryKey() {
			value := v.Field(column.Index())
			args = append(args, value.Interface())
		}
	}
	return args
}

// names_of returns the column names for the class, optionally excluding the
// names for any primary key.
func (this *sqclass) names_of(exclude_primary bool) []string {
	names := make([]string, 0, len(this.columns))
	for _, column := range this.columns {
		if (exclude_primary == false) || (column.PrimaryKey() == false) {
			names = append(names, column.Name())
		}
	}
	return names
}

// values_of returns the values for a struct, optionally excluding the
// values for any primary key.
func (this *sqclass) values_of(v reflect.Value, exclude_primary bool) []interface{} {
	if v.Kind() != reflect.Struct {
		return nil
	}
	args := make([]interface{}, 0, len(this.columns))
	for _, column := range this.columns {
		if (exclude_primary == false) || (column.PrimaryKey() == false) {
			value := v.Field(column.Index())
			args = append(args, value.Interface())
		}
	}
	return args
}

// query_expressions returns expressions for the SELECT statement with the keys
// first following by the remaining expressions
func (this *sqclass) query_expressions() []sq.Expression {
	expr := make([]sq.Expression, 0, len(this.columns))
	// Start with the keys
	for _, key := range this.keys {
		expr = append(expr, this.lang.NewSource(key))
	}
	// Append other columns
	for _, name := range this.names_of(this.object == false) {
		expr = append(expr, this.lang.NewSource(name))
	}
	return expr
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqclass) String() string {
	return fmt.Sprintf("<sqobj.Class>{ name=%v table_name=%v is_object=%v }", strconv.Quote(this.name), strconv.Quote(this.tablename), this.object)
}
