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
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sqlite "github.com/djthorpe/sqlite"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/sqlite/sys/sqlite"
	_ "github.com/djthorpe/sqlite/sys/sqobj"
)

////////////////////////////////////////////////////////////////////////////////

type Device struct {
	ID          uint      `sql:"device_id"`
	Name        string    `sql:"name"`
	DateAdded   time.Time `sql:"date_added"`
	DateUpdated time.Time `sql:"date_updated,nullable"`
	Enabled     bool      `sql:"enabled"`
}

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
		return gopi.ErrAppError
	} else if class, err := db.RegisterStruct("device", Device{}); err != nil {
		return err
	} else {
		fmt.Println(class)
	}

	// Success
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("db/sqobj")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
