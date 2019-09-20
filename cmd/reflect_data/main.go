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

	if db := app.ModuleInstance("db/sqlite").(sqlite.Connection); db == nil {
		return gopi.ErrAppError
	} else if columns, err := db.Reflect(Device{}); err != nil {
		return err
	} else if table := db.NewCreateTable("device", columns...); table == nil {
		return gopi.ErrBadParameter
	} else if _, err := db.Do(table); err != nil {
		return err
	} else if columns, err := db.ColumnsForTable("device", ""); err != nil {
		return err
	} else {
		fmt.Println(table.Query(db), columns)
	}

	// Success
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("db/sqlite")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
