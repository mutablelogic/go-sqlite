package sqobj

import (
	"fmt"

	// Import Namespaces
	. "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type sqiterator struct {
	proto interface{}
	rs    SQRows
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewIterator(proto interface{}, rs SQRows) *sqiterator {
	this := new(sqiterator)
	this.proto = proto
	this.rs = rs
	return this
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sqiterator) Next() interface{} {
	if this.proto == nil || this.rs == nil {
		return nil
	}
	if params := this.rs.NextArray(); params == nil {
		this.rs = nil
		return nil
	} else {
		fmt.Println("TODO:", params)
		return this.proto
	}
}

func (this *sqiterator) Close() error {
	if this.rs == nil {
		return nil
	} else {
		return this.rs.Close()
	}
}
