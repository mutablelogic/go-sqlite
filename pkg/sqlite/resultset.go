package sqlite

import (
	sql "database/sql/driver"

	// Modules

	sqlite "github.com/djthorpe/go-sqlite"
	driver "github.com/mattn/go-sqlite3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type resultset struct {
	r       *driver.SQLiteRows
	columns []string
	values  []sql.Value
	array   []interface{}
	zipped  map[string]interface{}
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewRows(r *driver.SQLiteRows) sqlite.SQRows {
	this := new(resultset)
	this.r = r
	this.columns = r.Columns()
	// pre-allocate values, array and map return values
	this.values = make([]sql.Value, len(this.columns))
	this.array = make([]interface{}, len(this.columns))
	this.zipped = make(map[string]interface{}, len(this.columns))
	// return success
	return this
}

func (this *resultset) Close() error {
	this.zipped = nil
	this.values = nil
	this.array = nil
	this.columns = nil
	return this.r.Close()
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *resultset) Next() []interface{} {
	if this.values == nil {
		return nil
	} else if err := this.r.Next(this.values); err != nil {
		this.Close()
		return nil
	} else if len(this.values) == 0 {
		this.Close()
		return nil
	} else {
		for i, v := range this.values {
			this.array[i] = v
		}
		return this.array
	}
}

func (this *resultset) NextMap() map[string]interface{} {
	if v := this.Next(); v == nil {
		return nil
	}
	for i, k := range this.columns {
		this.zipped[k] = this.values[i]
	}
	return this.zipped
}
