/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqlite

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	gopi.RegisterModule(gopi.Module{
		Name: "db/sqlite",
		Type: gopi.MODULE_TYPE_OTHER,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("sqlite.dsn", ":memory:", "Database source")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			dsn, _ := app.AppFlags.GetString("sqlite.dsn")
			return gopi.Open(Config{
				Path: dsn,
			}, app.Logger)
		},
	})
}
