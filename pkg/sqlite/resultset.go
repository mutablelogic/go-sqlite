package sqlite

import (
	sql "database/sql/driver"
	"io"

	// Modules
	marshaler "github.com/djthorpe/go-marshaler"
	sqlite "github.com/djthorpe/go-sqlite"
	driver "github.com/mattn/go-sqlite3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type resultset struct {
	r       *driver.SQLiteRows
	columns []string
	values  []sql.Value
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewRows(r *driver.SQLiteRows) sqlite.SQRows {
	this := new(resultset)
	if r == nil {
		return nil
	} else {
		this.r = r
		this.columns = r.Columns()
		this.values = make([]sql.Value, len(this.columns))
	}

	// Return success
	return this
}

func (this *resultset) Close() error {
	this.values = nil
	return this.r.Close()
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *resultset) Next(v interface{}) error {
	r := this.NextMap()
	if r == nil {
		return io.EOF
	} else if err := marshaler.UnmarshalStruct(r, v, sqlite.TagName, nil); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *resultset) NextMap() map[string]interface{} {
	v := this.NextArray()
	if v == nil {
		return nil
	}
	m := make(map[string]interface{}, len(v))
	for i := range v {
		m[this.columns[i]] = v[i]
	}
	return m
}

func (this *resultset) NextArray() []interface{} {
	if err := this.r.Next(this.values); err != nil {
		this.Close()
		return nil
	} else if len(this.values) == 0 {
		this.Close()
		return nil
	}
	r := make([]interface{}, len(this.values))
	for i := range this.values {
		r[i] = this.values[i]
	}
	return r
}
