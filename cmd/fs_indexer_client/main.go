/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"os"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, services []gopi.RPCServiceRecord, done chan<- struct{}) error {
	if indexer, err := app.ClientPool.NewClientEx("fsindexer.Indexer", services, gopi.RPC_FLAG_NONE); err != nil {
		return err
	} else if err := RunCommand(app, indexer.(sq.FSIndexerIndexClient)); err != nil {
		return err
	}

	// Success
	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/fsindexer/indexer:client")

	// Run the server and register all the services
	os.Exit(rpc.Client(config, time.Second*2, Main))
}
