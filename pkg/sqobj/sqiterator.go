package sqobj

import (
	"errors"
	"io"
	"reflect"

	// Modules

	// Import Namespaces
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Iterator struct {
	class *Class
	proto reflect.Value
	t     []reflect.Type
	rs    SQResults
	rowid int64
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func iterator(class *Class, rs SQResults) *Iterator {
	this := new(Iterator)

	// Set the class, prototype object and results
	this.class = class
	this.proto = class.Proto()
	this.rs = rs

	// Set the casting types - first is the rowid, then the rest are the values
	this.t = append(this.t, reflect.TypeOf(int64(0)))
	for _, col := range this.class.col {
		this.t = append(this.t, col.Type)
	}

	return this
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (i *Iterator) Next() interface{} {
	if i.rs == nil {
		return nil
	}
	v, err := i.rs.Next(i.t...)
	if err != nil {
		i.rs = nil
		i.rowid = 0
		if !errors.Is(err, io.EOF) {
			panic(err)
		}
		return nil
	}

	// Set rowid and proto values
	i.rowid = v[0].(int64)
	i.class.unboundValues(i.proto, v[1:])

	// Return the prototype object
	return i.proto.Interface()
}

func (i *Iterator) RowId() int64 {
	return i.rowid
}
