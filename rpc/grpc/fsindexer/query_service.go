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
	sq "github.com/djthorpe/sqlite"

	// Protocol buffers
	pb "github.com/djthorpe/sqlite/rpc/protobuf/fsindexer"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type QueryService struct {
	Server  gopi.RPCServer
	Indexer sq.FSIndexer
}

type query_service struct {
	log     gopi.Logger
	indexer sq.FSIndexer
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config QueryService) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.fsindexer.query>Open{ config=%+v }", config)

	this := new(query_service)
	this.log = log
	if config.Indexer == nil {
		return nil, gopi.ErrBadParameter
	} else {
		this.indexer = config.Indexer
	}

	// Register service with GRPC server
	pb.RegisterQueryServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return this, nil
}

func (this *query_service) Close() error {
	this.log.Debug("<grpc.service.fsindexer.query>Close{}")

	// Release resources
	this.indexer = nil

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
	return fmt.Sprintf("grpc.service.fsindexer.query{ %v }", this.indexer)
}

////////////////////////////////////////////////////////////////////////////////
// RPC Methods

func (this *query_service) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.fsindexer.query.Ping>{ }")
	return &empty.Empty{}, nil
}

func (this *query_service) List(context.Context, *empty.Empty) (*pb.ListResponse, error) {
	return nil, gopi.ErrNotImplemented
}
