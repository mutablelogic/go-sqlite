package ncruces

import (
	// Packages
	values "github.com/mutablelogic/go-sqlite/values"
	sqlite3 "github.com/ncruces/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// column is a row's column, which is a values.Value
type column struct {
	stmt *sqlite3.Stmt
	i    int
}

// param is a statement's argument, which is a values.Setter
type param struct {
	stmt *sqlite3.Stmt
	i    int
}

// arg is a function's argument, which is a values.Value
type arg struct {
	value sqlite3.Value
}

// result is a function's result, which is a values.Setter
type result struct {
	ctx sqlite3.Context
}

// Ensure interfaces are satisfied
var _ values.Value = column{}
var _ values.Value = arg{}
var _ values.Setter = param{}
var _ values.Setter = result{}

////////////////////////////////////////////////////////////////////////////////
// COLUMN

func (c column) Type() values.Type { return valueType(c.stmt.ColumnType(c.i)) }
func (c column) Int64() int64      { return c.stmt.ColumnInt64(c.i) }
func (c column) Float() float64    { return c.stmt.ColumnFloat(c.i) }
func (c column) Text() string      { return c.stmt.ColumnText(c.i) }
func (c column) Blob() []byte      { return c.stmt.ColumnBlob(c.i, nil) }

////////////////////////////////////////////////////////////////////////////////
// PARAM

func (p param) SetNull() error           { return p.stmt.BindNull(p.i) }
func (p param) SetInt64(n int64) error   { return p.stmt.BindInt64(p.i, n) }
func (p param) SetFloat(f float64) error { return p.stmt.BindFloat(p.i, f) }
func (p param) SetText(s string) error   { return p.stmt.BindText(p.i, s) }
func (p param) SetBlob(b []byte) error   { return p.stmt.BindBlob(p.i, b) }

////////////////////////////////////////////////////////////////////////////////
// ARG

func (a arg) Type() values.Type { return valueType(a.value.Type()) }
func (a arg) Int64() int64      { return a.value.Int64() }
func (a arg) Float() float64    { return a.value.Float() }
func (a arg) Text() string      { return a.value.Text() }
func (a arg) Blob() []byte      { return a.value.Blob(nil) }

////////////////////////////////////////////////////////////////////////////////
// RESULT

func (r result) SetNull() error           { r.ctx.ResultNull(); return nil }
func (r result) SetInt64(n int64) error   { r.ctx.ResultInt64(n); return nil }
func (r result) SetFloat(f float64) error { r.ctx.ResultFloat(f); return nil }
func (r result) SetText(s string) error   { r.ctx.ResultText(s); return nil }
func (r result) SetBlob(b []byte) error   { r.ctx.ResultBlob(b); return nil }

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// valueType returns SQLite's type as a values.Type
func valueType(t sqlite3.Datatype) values.Type {
	switch t {
	case sqlite3.INTEGER:
		return values.Integer
	case sqlite3.FLOAT:
		return values.Float
	case sqlite3.TEXT:
		return values.Text
	case sqlite3.BLOB:
		return values.Blob
	}
	return values.Null
}
