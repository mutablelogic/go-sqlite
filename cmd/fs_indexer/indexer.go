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

func (this *Indexer) Do(file *File) error {
	if file == nil || file.Id == 0 || file.Path == "" || file.Root == "" {
		return gopi.ErrBadParameter
	} else if path, err := filepath.Rel(file.Root, file.Path); err != nil {
		return err
	} else if _, err := this.sqobj.Insert(&File{file.Id, path, file.Root}); err != nil {
		return err
	} else {
		return nil
	}
}
