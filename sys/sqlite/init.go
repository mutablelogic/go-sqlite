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
			config.AppFlags.FlagString("sqlite.tz", "", "Timezone for parsed timestamps")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			dsn, _ := app.AppFlags.GetString("sqlite.dsn")
			tz, _ := app.AppFlags.GetString("sqlite.tz")
			return gopi.Open(Database{
				Path:     dsn,
				Location: tz,
			}, app.Logger)
		},
	})

	gopi.RegisterModule(gopi.Module{
		Name: "db/sqlang",
		Type: gopi.MODULE_TYPE_OTHER,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Language{}, app.Logger)
		},
	})
}
