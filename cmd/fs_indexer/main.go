/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"fmt"
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
	Id   int64  `sql:"id,primary"`
	Root string `sql:"root,primary"`
	Path string `sql:"path"`
	sqlite.Object
}

var (
	file_chan = make(chan *File, 0)
)

func Index(app *gopi.AppInstance, folder string) {
	if err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Ignore hidden folders
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		} else if info.Mode().IsRegular() == false {
			// Ignore files which aren't regular files
			return nil
		} else if strings.HasPrefix(info.Name(), ".") {
			// Ignore hidden files
		} else if inode := inodeForInfo(info); inode == 0 {
			// inode not obtained
		} else {
			file_chan <- &File{inode, folder, path, sqlite.Object{}}
		}
		return nil
	}); err != nil {
		app.Logger.Error("Error: %v", err)
	} else {
		app.Logger.Info("Indexed: %v", folder)
	}
}

func IndexFile(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	sqobj := app.ModuleInstance("db/sqobj").(sqlite.Objects)
	indexer := NewIndexer(sqobj)
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

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if len(app.AppFlags.Args()) == 0 {
		return gopi.ErrHelp
	} else {
		for _, folder := range app.AppFlags.Args() {
			name := strings.TrimSuffix(filepath.Base(folder), filepath.Ext(folder))
			if s, err := os.Stat(folder); os.IsNotExist(err) {
				return fmt.Errorf("%v: Does not exist", name)
			} else if err != nil {
				return fmt.Errorf("%v: %v", name, err)
			} else if s.Mode().IsDir() == false {
				return fmt.Errorf("%v: Not a folder", name)
			} else {
				go Index(app, filepath.Clean(folder))
			}
		}
	}

	// Wait for interrupt
	app.Logger.Info("Press CTRL+C to cancel")
	app.WaitForSignal()

	// Success
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("db/sqobj")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, IndexFile))
}
