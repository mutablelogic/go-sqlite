/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

/*
import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sqlite "github.com/djthorpe/sqlite"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/sqlite/sys/sqlite"
	_ "github.com/djthorpe/sqlite/sys/sqobj"
)

type File struct {
	Id       int64  `sql:"id,primary"`
	Root     string `sql:"root,primary"`
	Path     string `sql:"path"`
	MimeType string `sql:"mime_type"`
	Size     int64  `sql:"size"`
	sqlite.Object
}

var (
	// file_chan is the channel used for indexing files
	file_chan = make(chan *File, 0)
)

// Index will kick-off an indexing task
func Index(app *gopi.AppInstance, folder string) {
	if err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		// If not readable or executable folder, then ignore
		if info.IsDir() && isReadableFileAtPath(path) != nil {
			return filepath.SkipDir
		}
		if info.IsDir() && isExecutableFileAtPath(path) != nil {
			return filepath.SkipDir
		}
		// Return any other errors
		if err != nil {
			return err
		}
		// Ignore hidden folders
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		} else if info.Mode().IsRegular() == false {
			// Ignore files which aren't regular files
		} else if strings.HasPrefix(info.Name(), ".") {
			// Ignore hidden files
		} else if inode := inodeForInfo(info); inode == 0 {
			// inode not obtained
		} else if mimetype, err := DetectMimeType(path); err != nil {
			app.Logger.Error("Error: %v", err)
		} else {
			file_chan <- &File{inode, folder, path, mimetype, info.Size(), sqlite.Object{}}
		}
		// Return success - continue
		return nil
	}); err != nil {
		app.Logger.Error("Error: %v", err)
	}
}

// IndexFile is a go-routine for placing a file into the database
func IndexFile(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	sqobj := app.ModuleInstance("db/sqobj").(sqlite.Objects)
	indexer, err := NewIndexer(sqobj)
	if err != nil {
		return err
	}
	timer := time.NewTicker(2 * time.Second)
	start <- gopi.DONE
	total_rows, total_rows_ := uint64(0), uint64(0)

FOR_LOOP:
	for {
		select {
		case file := <-file_chan:
			if affected_rows, err := indexer.Do(file); err != nil {
				app.Logger.Error("Error: %v: %v", filepath.Base(file.Path), err)
			} else {
				total_rows += affected_rows
			}
		case <-timer.C:
			if total_rows_ != total_rows {
				app.Logger.Info("%v files indexed", total_rows)
				total_rows_ = total_rows
			}
		case <-stop:
			timer.Stop()
			break FOR_LOOP
		}
	}
	return nil
}
*/
