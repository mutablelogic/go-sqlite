/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package sqobj

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sqlite "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	gopi.RegisterModule(gopi.Module{
		Name:     "db/sqobj",
		Type:     gopi.MODULE_TYPE_OTHER,
		Requires: []string{"db/sqlite"},
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagBool("sqlite.create", true, "Create tables")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			create, _ := app.AppFlags.GetBool("sqlite.create")
			return gopi.Open(Config{
				Conn:   app.ModuleInstance("db/sqlite").(sqlite.Connection),
				Create: create,
			}, app.Logger)
		},
	})
}
