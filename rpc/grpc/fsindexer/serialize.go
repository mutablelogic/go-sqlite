/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package fsindexer

import (
	// Frameworks
	"fmt"
	"strconv"

	sq "github.com/djthorpe/sqlite"

	// Protocol buffers
	pb "github.com/djthorpe/sqlite/rpc/protobuf/fsindexer"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type fsindex_proto struct {
	*pb.Index
}

////////////////////////////////////////////////////////////////////////////////
// FSINDEX IMPLEMENTATION

func (this *fsindex_proto) Id() int64 {
	if this.Index == nil {
		return 0
	} else {
		return this.Index.Id
	}
}

func (this *fsindex_proto) Name() string {
	if this.Index == nil {
		return ""
	} else {
		return this.Index.Name
	}
}

func (this *fsindex_proto) Count() uint64 {
	if this.Index == nil {
		return 0
	} else {
		return this.Index.Count
	}
}

func (this *fsindex_proto) Status() sq.FSStatus {
	if this.Index == nil {
		return sq.FS_STATUS_NONE
	} else {
		return sq.FSStatus(this.Index.Status)
	}
}

func (this *fsindex_proto) String() string {
	return fmt.Sprintf("<FSIndex>{ id=%v name=%v count=%v status=%v }", this.Id(), strconv.Quote(this.Name()), this.Count(), this.Status())
}

////////////////////////////////////////////////////////////////////////////////
// LIST OF FSINDEX

func to_fsindex_proto(indexes []sq.FSIndex) []*pb.Index {
	if indexes == nil {
		return nil
	}
	proto := make([]*pb.Index, len(indexes))
	for i, index := range indexes {
		proto[i] = &pb.Index{
			Id:     index.Id(),
			Name:   index.Name(),
			Count:  index.Count(),
			Status: pb.Index_IndexStatus(index.Status()),
		}
	}
	return proto
}
