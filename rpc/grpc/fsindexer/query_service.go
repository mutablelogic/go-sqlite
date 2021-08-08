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
// CONSTANTS

const (
	QUERY_LIMIT_DEFAULT = 50
	QUERY_LIMIT_MAX     = 1000
)

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
	this.log.Debug("<grpc.service.fsindexer.query.List>{ }")
	return &pb.ListResponse{
		Index: to_fsindex_proto(this.indexer.Indexes()),
	}, nil
}

func (this *query_service) Query(_ context.Context, req *pb.QueryRequest) (*pb.QueryResponse, error) {
	this.log.Debug("<grpc.service.fsindexer.query.Query>{ req=%v }", req)

	// Check incoming parameters
	if req.Limit == 0 {
		req.Limit = QUERY_LIMIT_DEFAULT
	} else if req.Limit > QUERY_LIMIT_MAX {
		return nil, fmt.Errorf("%w: Maximum limit is %v", gopi.ErrBadParameter, QUERY_LIMIT_MAX)
	}

	// Perform count
	if count, err := this.indexer.Count(); err != nil {
		return nil, err
	} else if files, err := this.indexer.Query(req.Limit); err != nil {
		return nil, err
	} else {
		return &pb.QueryResponse{
			Count: count,
			Limit: req.Limit,
			File:  to_fsfile_proto(files),
		}, nil
	}
}
