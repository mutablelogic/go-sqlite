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
	// Query Service
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/fsindexer/query:service",
		Type:     gopi.MODULE_TYPE_SERVICE,
		Requires: []string{"rpc/server", "db/fsindexer"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(QueryService{
				Server:  app.ModuleInstance("rpc/server").(gopi.RPCServer),
				Indexer: app.ModuleInstance("db/fsindexer").(sq.FSIndexer),
			}, app.Logger)
		},
	})

	// Indexer Service
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/fsindexer/indexer:service",
		Type:     gopi.MODULE_TYPE_SERVICE,
		Requires: []string{"rpc/server", "db/fsindexer"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(IndexerService{
				Server:  app.ModuleInstance("rpc/server").(gopi.RPCServer),
				Indexer: app.ModuleInstance("db/fsindexer").(sq.FSIndexer),
			}, app.Logger)
		},
	})

	// Indexer Client
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/fsindexer/client",
		Type:     gopi.MODULE_TYPE_CLIENT,
		Requires: []string{"rpc/clientpool"},
		Run: func(app *gopi.AppInstance, _ gopi.Driver) error {
			if clientpool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool); clientpool == nil {
				return gopi.ErrAppError
			} else {
				clientpool.RegisterClient("fsindexer.Indexer", NewIndexerClient)
				clientpool.RegisterClient("fsindexer.Query", NewQueryClient)
				return nil
			}
		},
	})

}
