/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"fmt"
	"path/filepath"

	// Frameworks
	sqlite "github.com/djthorpe/sqlite"
)

type Indexer struct {
	sqobj sqlite.Objects
}

type File struct {
	Root string
	Path string
}

func NewIndexer(sqobj sqlite.Objects) *Indexer {
	this := new(Indexer)
	this.sqobj = sqobj

	if _, err := sqobj.RegisterStruct(&File{}); err != nil {
		return nil
	}

	return this
}

func (this *Indexer) Do(path, root string) error {
	if path, err := filepath.Rel(root, path); err != nil {
		return err
	} else {
		fmt.Println(path, "=>", root)
	}
	return nil
}
