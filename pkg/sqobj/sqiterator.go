package sqobj

import (
	// Import Namespaces
	. "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqiterator struct {
	class *sqclass
	rs    SQRows
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewIterator(class *sqclass, rs SQRows) *sqiterator {
	this := new(sqiterator)
	this.class = class
	this.rs = rs
	return this
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqiterator) Next() interface{} {
	if this.rs == nil {
		return nil
	}
	if params := this.rs.Next(); params == nil {
		this.rs = nil
		return nil
	} else if obj, err := this.class.Object(params); err != nil {
		panic(err)
	} else {
		return obj
	}
}

func (this *sqiterator) Close() error {
	if this.rs == nil {
		return nil
	} else {
		return this.rs.Close()
	}
}
