/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sqlite "github.com/djthorpe/sqlite"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/sqlite/sys/sqlite"
)

func Process(app *gopi.AppInstance, db sqlite.Connection, name string, fh io.Reader) error {
	reader := csv.NewReader(fh)
	noheader, _ := app.AppFlags.GetBool("noheader")
	table := NewColumns(name)

	// Scan table for name and types
	for i := 0; true; i++ {
		if row, err := reader.Read(); err == io.EOF {
			break
		} else if err != nil {
			return err
		} else if i == 0 && noheader == false {
			table.SetNames(row)
		} else {
			table.SetTypes(row)
		}
	}

	// Print out table
	fmt.Println(table)

	// Return success
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	if db := app.ModuleInstance("db/sqlite").(sqlite.Connection); db == nil {
		return gopi.ErrAppError
	} else if len(app.AppFlags.Args()) == 0 {
		return gopi.ErrHelp
	} else {
		for _, filename := range app.AppFlags.Args() {
			name := filepath.Base(filename)
			if s, err := os.Stat(filename); err != nil {
				return fmt.Errorf("%v: %v", name, err)
			} else if s.Mode().IsRegular() == false {
				return fmt.Errorf("%v: Not a regular file", name)
			} else if fh, err := os.Open(filename); err != nil {
				return err
			} else {
				defer fh.Close()
				if err := Process(app, db, name, fh); err != nil {
					return fmt.Errorf("%v: %v", name, err)
				}
			}
		}
	}

	// Success
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("db/sqlite")

	// Set arguments
	config.AppFlags.FlagBool("noheader", false, "Do not use the first row as column names")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
