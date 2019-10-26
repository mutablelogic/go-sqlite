/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"os"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	fsindexer := app.ModuleInstance("db/fsindexer").(sq.FSIndexer)
	if len(app.AppFlags.Args()) == 0 {
		return gopi.ErrHelp
	}
	for _, folder := range app.AppFlags.Args() {
		if _, err := fsindexer.Index(folder, false); err != nil {
			return err
		}
	}
	/*
		name := strings.TrimSuffix(filepath.Base(folder), filepath.Ext(folder))
		if s, err := os.Stat(folder); os.IsNotExist(err) {
			return fmt.Errorf("%v: Does not exist", name)
		} else if err != nil {
			return fmt.Errorf("%v: %v", name, err)
		} else if s.Mode().IsDir() == false {
			return fmt.Errorf("%v: Not a folder", name)
		} else {
			go Index(app, filepath.Clean(folder))
		}*/

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
	config := gopi.NewAppConfig("db/fsindexer")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
