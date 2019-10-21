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

type IndexService struct {
	Server  gopi.RPCServer
	Indexer sq.FSIndexer
}

type index_service struct {
	log gopi.Logger
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config IndexService) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.fsindexer.index>Open{ server=%v }", config.Server)

	this := new(index_service)
	this.log = log

	// Register service with GRPC server
	pb.RegisterIndexServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return this, nil
}

func (this *index_service) Close() error {
	this.log.Debug("<grpc.service.fsindexer.index>Close{}")

	// No resources to release

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// RPCService implementation

func (this *index_service) CancelRequests() error {
	// No need to cancel any requests since none are streaming
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Stringify

func (this *index_service) String() string {
	return fmt.Sprintf("grpc.service.fsindexer.index{}")
}

////////////////////////////////////////////////////////////////////////////////
// RPC Methods

func (this *index_service) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.fsindexer.index.Ping>{ }")
	return &empty.Empty{}, nil
}

func (this *index_service) Index(ctx context.Context, req *pb.IndexRequest) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.fsindexer.index.Index>{ req=%v }", req)

	// Perform the index of a volume

	return &empty.Empty{}, nil
}
