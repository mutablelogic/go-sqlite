package main

import (
	"fmt"
	"time"

	// Modules
	sq "github.com/djthorpe/go-sqlite"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
	sqobj "github.com/djthorpe/go-sqlite/pkg/sqobj"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type File struct {
	Path    string    `sqlite:"path,primary"`
	Name    string    `sqlite:"name,index:name"`
	ModTime time.Time `sqlite:"modtime"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tFile = "file"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func OpenDatabase(filename string) (sq.SQConnection, error) {
	// Open database
	db, err := sqlite.Open(filename, time.Local)
	if err != nil {
		return nil, err
	}

	if err := db.Do(func(tx sq.SQTransaction) error {
		// Create tables and indexes
		q := sqobj.CreateTable(tFile, File{}).IfNotExists()
		if _, err := tx.Exec(q); err != nil {
			return err
		}
		for _, q := range sqobj.CreateIndexes(tFile, File{}) {
			q = q.IfNotExists()
			if _, err := tx.Exec(q); err != nil {
				return err
			}
		}
		// Return success
		return nil
	}); err != nil {
		return nil, err
	}

	// Return success
	return db, nil
}

func InsertRow(db sq.SQConnection, file File) error {
	q := sqobj.InsertRow(tFile, file)
	return db.Do(func(tx sq.SQTransaction) error {
		fmt.Println(q)
		// Return success
		return nil
	})
}
