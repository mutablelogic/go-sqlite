/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package fsindexer

import (
	"fmt"
	"strconv"

	// Frameworks
	sq "github.com/djthorpe/sqlite"

	// Protocol buffers
	pb "github.com/djthorpe/sqlite/rpc/protobuf/fsindexer"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type fsindex_proto struct {
	*pb.Index
}

type fsquery_proto struct {
	*pb.QueryResponse
}

type fsfile_proto struct {
	*pb.File
}

////////////////////////////////////////////////////////////////////////////////
// FSFILE IMPLEMENTATION

func NewFile(pb *pb.File) sq.FSFile {
	if pb == nil {
		return nil
	} else {
		return &fsfile_proto{pb}
	}
}

func (this *fsfile_proto) Id() int64 {
	return this.File.Id
}
func (this *fsfile_proto) Index() sq.FSIndex {
	return NewIndex(this.File.Index)
}

func (this *fsfile_proto) Path() string {
	return this.File.Path
}

func (this *fsfile_proto) Name() string {
	return this.File.Name
}

func (this *fsfile_proto) Ext() string {
	return this.File.Ext
}

func (this *fsfile_proto) MimeType() string {
	return this.File.Mimetype
}

func (this *fsfile_proto) Size() int64 {
	return this.File.Size
}

////////////////////////////////////////////////////////////////////////////////
// FSINDEX IMPLEMENTATION

func NewIndex(pb *pb.Index) sq.FSIndex {
	if pb == nil {
		return nil
	} else {
		return &fsindex_proto{pb}
	}
}

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
// FSQUERY IMPLEMENTATION

func NewQueryResponse(pb *pb.QueryResponse) sq.FSQueryResponse {
	return &fsquery_proto{pb}
}

func (this *fsquery_proto) Count() uint64 {
	if this.QueryResponse == nil {
		return 0
	} else {
		return this.QueryResponse.Count
	}
}

func (this *fsquery_proto) Limit() uint64 {
	if this.QueryResponse == nil {
		return 0
	} else {
		return this.QueryResponse.Limit
	}
}

func (this *fsquery_proto) Files() []sq.FSFile {
	if this.QueryResponse == nil {
		return nil
	} else {
		files := make([]sq.FSFile, len(this.QueryResponse.File))
		for i, file := range this.QueryResponse.File {
			files[i] = NewFile(file)
		}
		return files
	}
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

////////////////////////////////////////////////////////////////////////////////
// LIST OF FSFILE

func to_fsfile_proto(files []sq.FSFile) []*pb.File {
	if files == nil {
		return nil
	}
	proto := make([]*pb.File, len(files))
	for i, file := range files {
		proto[i] = &pb.File{
			Id:   file.Id(),
			Name: file.Name(),
		}
	}
	return proto
}
