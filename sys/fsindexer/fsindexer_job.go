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
)

////////////////////////////////////////////////////////////////////////////////
// FSIndex implementation

func (this *job) Id() int64 {
	return this.jobnode
}

func (this *job) Name() string {
	return this.relpath
}

func (this *job) Count() uint64 {
	return this.count
}

func (this *job) Status() sq.FSStatus {
	this.Lock()
	defer this.Unlock()
	if this.done {
		return sq.FS_STATUS_IDLE
	} else {
		return sq.FS_STATUS_INDEXING
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *job) String() string {
	return fmt.Sprintf("<FSIndex>{ id=%v name=%v count=%v status=%v }", this.Id(), strconv.Quote(this.Name()), this.Count(), this.Status())
}
