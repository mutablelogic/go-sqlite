/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package fsindexer

import (
	// Frameworks

	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// QUUERY IMPLEMENTATION

func (this *indexer) Count() (uint64, error) {
	this.log.Debug2("<fsindexer.Count>{ }")
	if class := this.sqobj.ClassFor(&File{}); class == nil {
		return 0, gopi.ErrAppError
	} else if count, err := this.sqobj.Count(class); err != nil {
		return 0, err
	} else {
		return count, nil
	}
}

func (this *indexer) Query(limit uint64) ([]sq.FSFile, error) {
	this.log.Debug2("<fsindexer.Query>{ limit=%v }", limit)
	// Check parameters
	if limit == 0 {
		return nil, gopi.ErrBadParameter
	}
	// Make array with enough capacity
	files := make([]File, 0, limit)
	if _, err := this.sqobj.Read(&files, uint(limit)); err != nil {
		return nil, err
	} else if len(files) == 0 {
		return []sq.FSFile{}, nil
	} else {
		fsfiles := make([]sq.FSFile, len(files))
		for i, file := range files {
			fsfiles[i] = NewFSFile(&file)
		}
		return fsfiles, nil
	}
}
