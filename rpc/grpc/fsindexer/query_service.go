/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package fsindexer

import (
	"context"
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"

	// Protocol buffers
	pb "github.com/djthorpe/sqlite/rpc/protobuf/fsindexer"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type QueryService struct {
	Server gopi.RPCServer
}

type query_service struct {
	log gopi.Logger
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config QueryService) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.fsindexer.query>Open{ server=%v }", config.Server)

	this := new(query_service)
	this.log = log

	// Register service with GRPC server
	pb.RegisterQueryServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return this, nil
}

func (this *query_service) Close() error {
	this.log.Debug("<grpc.service.fsindexer.query>Close{}")

	// No resources to release

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// RPCService implementation

func (this *query_service) CancelRequests() error {
	// No need to cancel any requests since none are streaming
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Stringify

func (this *query_service) String() string {
	return fmt.Sprintf("grpc.service.fsindexer.query{}")
}

////////////////////////////////////////////////////////////////////////////////
// RPC Methods

func (this *query_service) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.fsindexer.query.Ping>{ }")
	return &empty.Empty{}, nil
}
