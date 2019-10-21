/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package fsindexer

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	gopi.RegisterModule(gopi.Module{
		Name:     "db/fsindexer",
		Type:     gopi.MODULE_TYPE_OTHER,
		Requires: []string{"db/sqobj"},
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("fsindexer.root", "", "Root path for the indexer")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			root, _ := app.AppFlags.GetString("fsindexer.root")
			return gopi.Open(Indexer{
				Root:    root,
				Objects: app.ModuleInstance("db/sqobj").(sq.Objects),
			}, app.Logger)
		},
	})
}
