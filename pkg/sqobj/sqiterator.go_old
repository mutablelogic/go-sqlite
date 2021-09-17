package sqobj

import (
	"reflect"

	// Modules

	// Import Namespaces
	. "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqiterator struct {
	class *sqclass
	proto reflect.Value
	rs    SQRows
	rowid int64
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewIterator(class *sqclass, rs SQRows) *sqiterator {
	this := new(sqiterator)
	this.class = class
	this.proto = class.Proto()
	this.rs = rs
	return this
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqiterator) Next() interface{} {
	if this.rs == nil {
		return nil
	}
	params := this.rs.Next()
	if params == nil {
		this.rs = nil
		this.rowid = 0
		return nil
	}
	if err := this.class.unboundValues(this.proto, params[1:]); err != nil {
		panic(err)
	}
	this.rowid = params[0].(int64)
	return this.proto
}

func (this *sqiterator) RowId() int64 {
	return this.rowid
}

func (this *sqiterator) Close() error {
	if this.rs == nil {
		return nil
	} else {
		return this.rs.Close()
	}
}
