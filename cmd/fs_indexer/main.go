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

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sqlite "github.com/djthorpe/sqlite"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/sqlite/sys/sqlite"
)

func Index(app *gopi.AppInstance, indexer *Indexer, folder string) {
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
		} else if err := indexer.Do(path, folder); err != nil {
			return err
		}
		return nil
	}); err != nil {
		app.Logger.Error("Error: %v", err)
	}
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if len(app.AppFlags.Args()) == 0 {
		return gopi.ErrHelp
	} else {
		sqobj := app.ModuleInstance("db/sqobj").(sqlite.Objects)
		indexer := NewIndexer(sqobj)
		for _, folder := range app.AppFlags.Args() {
			name := strings.TrimSuffix(filepath.Base(folder), filepath.Ext(folder))
			if s, err := os.Stat(folder); os.IsNotExist(err) {
				return fmt.Errorf("%v: Does not exist", name)
			} else if err != nil {
				return fmt.Errorf("%v: %v", name, err)
			} else if s.Mode().IsDir() == false {
				return fmt.Errorf("%v: Not a folder", name)
			} else {
				go Index(app, indexer, filepath.Clean(folder))
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
	config := gopi.NewAppConfig("db/sqlite", "db/sqlang")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
