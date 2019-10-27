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
	"path/filepath"
	"strconv"

	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type fsfile struct {
	*File
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func NewFSFile(file *File) sq.FSFile {
	if file == nil {
		return nil
	} else {
		return &fsfile{file}
	}
}

func (this *fsfile) Id() int64 {
	return this.File.Id
}

func (this *fsfile) Index() sq.FSIndex {
	// TODO
	return nil
}

func (this *fsfile) Path() string {
	return this.File.RelPath
}

func (this *fsfile) Name() string {
	return filepath.Base(this.File.RelPath)
}

func (this *fsfile) Ext() string {
	return this.File.Ext
}

func (this *fsfile) MimeType() string {
	return this.File.MimeType
}

func (this *fsfile) Size() int64 {
	return this.File.Size
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *fsfile) String() string {
	return fmt.Sprintf("<FSFile>{ id=%v index=%v path=%v name=%v ext=%v mimetype=%v size=%v }",
		this.Id(), this.Index(), strconv.Quote(this.Path()), strconv.Quote(this.Name()), strconv.Quote(this.Ext()),
		strconv.Quote(this.MimeType()), this.Size())

}
