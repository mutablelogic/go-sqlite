/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"path/filepath"

	// Frameworks
	sqlite "github.com/djthorpe/sqlite"
)

type Indexer struct {
	sqobj sqlite.Objects
}

type File struct {
	Root string `sql:"root"`
	Path string `sql:"path"`
}

func NewIndexer(sqobj sqlite.Objects) *Indexer {
	this := new(Indexer)
	this.sqobj = sqobj

	if _, err := sqobj.RegisterStruct(&File{}); err != nil {
		return nil
	}

	return this
}

func (this *Indexer) Do(root, path string) error {
	if path, err := filepath.Rel(root, path); err != nil {
		return err
	} else if _, err := this.sqobj.Insert(&File{root, path}); err != nil {
		return err
	}
	return nil
}
