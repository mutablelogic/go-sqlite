/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"path/filepath"

	"github.com/djthorpe/gopi"

	// Frameworks
	sqlite "github.com/djthorpe/sqlite"
)

type Indexer struct {
	sqobj sqlite.Objects
}

func NewIndexer(sqobj sqlite.Objects) *Indexer {
	this := new(Indexer)
	this.sqobj = sqobj
	if _, err := sqobj.RegisterStruct(&File{}); err != nil {
		return nil
	}
	return this
}

func (this *Indexer) Do(file *File) (uint64, error) {
	if file == nil || file.Id == 0 || file.Path == "" || file.Root == "" {
		return 0, gopi.ErrBadParameter
	} else if path, err := filepath.Rel(file.Root, file.Path); err != nil {
		return 0, err
	} else {
		file.Path = path
		if affected_rows, err := this.sqobj.Write(sqlite.FLAG_INSERT|sqlite.FLAG_UPDATE, file); err != nil {
			return 0, err
		} else {
			return affected_rows, nil
		}
	}
}
