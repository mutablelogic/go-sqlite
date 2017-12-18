// Connect to a database file
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	// Interfaces
	gopi "github.com/djthorpe/gopi"
	sqlite "github.com/djthorpe/sqlite"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	v3 "github.com/djthorpe/sqlite/v3"
)

// Basic information about an IoT "thing"
type IoTDevice struct {
	DeviceKey    string    `sql:"name:key;type:CHAR(20);unique;not null;"`
	Manufacturer string    `sql:"name:manufacturer;type:CHAR(80);"`
	ProductName  string    `sql:"name:product;type:CHAR(80);"`
	Description  string    `sql:"name:description;type:CHAR(100);"`
	Active       bool      `sql:"name:active;"`
	Paired       bool      `sql:"name:paired;"`
	TimeActive   time.Time `sql:"name:time_active;"`
	TimePaired   time.Time `sql:"name:time_paired;"`
	TimeUnpaired time.Time `sql:"name:time_unpaired;"`
	TimeUpdated  time.Time `sql:"name:time_updated;"`
	Duration     time.Duration
	Blob         []byte
}

////////////////////////////////////////////////////////////////////////////////

func RunLoop2(app *gopi.AppInstance, db sqlite.Client) error {
	app.Logger.Info("db=%v", db)

	// Reflection on the columns
	var device IoTDevice
	if columns, err := db.Reflect(&device); err != nil {
		return err
	} else {
		sql := sqlite.CreateTable("device", columns).IfNotExists()
		if err := db.Do(sql); err != nil {
			return err
		}
	}

	return nil
}

func RunLoop(app *gopi.AppInstance, done chan struct{}) error {

	config := v3.Client{}

	if dsn, exists := app.AppFlags.GetString("dsn"); exists == false {
		return errors.New("Missing -dsn flag")
	} else {
		config.DSN = dsn
	}

	// Create a client
	if client, err := gopi.Open(config, app.Logger); err != nil {
		return err
	} else {
		defer client.Close()
		if err := RunLoop2(app, client.(sqlite.Client)); err != nil {
			return err
		}
	}

	// Successful completion
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BOOTSTRAP THE APPLICATION

func registerFlags(config gopi.AppConfig) gopi.AppConfig {
	// Register the flags & return the configuration
	config.AppFlags.FlagString("dsn", "", "SQLite connection string")

	return config
}

func main_inner() int {
	// Set application configuration
	config := gopi.NewAppConfig()
	// Create the application with an empty configuration
	app, err := gopi.NewAppInstance(registerFlags(config))
	if err != nil {
		if err != gopi.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return -1
		}
		return 0
	}
	defer app.Close()

	// Run the application
	if err := app.Run(RunLoop); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return -1
	}
	return 0
}

func main() {
	os.Exit(main_inner())
}
