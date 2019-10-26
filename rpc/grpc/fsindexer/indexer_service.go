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

type IndexerService struct {
	Server  gopi.RPCServer
	Indexer sq.FSIndexer
}

type index_service struct {
	log     gopi.Logger
	indexer sq.FSIndexer
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config IndexerService) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.fsindexer.indexer>Open{ config=%+v }", config)

	this := new(index_service)
	this.log = log
	if config.Indexer == nil {
		return nil, gopi.ErrBadParameter
	} else {
		this.indexer = config.Indexer
	}

	// Register service with GRPC server
	pb.RegisterIndexerServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return this, nil
}

func (this *index_service) Close() error {
	this.log.Debug("<grpc.service.fsindexer.indexer>Close{ %v }", this.indexer)

	// Release resources
	this.indexer = nil

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
	return fmt.Sprintf("grpc.service.fsindexer.indexer{ %v }", this.indexer)
}

////////////////////////////////////////////////////////////////////////////////
// RPC Methods

func (this *index_service) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.fsindexer.indexer.Ping>{ }")
	return &empty.Empty{}, nil
}

func (this *index_service) List(context.Context, *empty.Empty) (*pb.ListResponse, error) {
	this.log.Debug("<grpc.service.fsindexer.indexer.List>{ }")

	// Return jobs
	return &pb.ListResponse{
		Index: to_fsindex_proto(this.indexer.Indexes()),
	}, nil
}

func (this *index_service) Index(ctx context.Context, req *pb.IndexRequest) (*pb.ListResponse, error) {
	this.log.Debug("<grpc.service.fsindexer.indexer.Index>{ req=%v }", req)

	// Perform the index of a volume
	if job, err := this.indexer.Index(req.Path, req.Watch); err != nil {
		return nil, err
	} else if job_ := this.indexer.IndexById(job); job_ == nil {
		return nil, gopi.ErrAppError
	} else {
		// Return a single response with the job id and name
		return &pb.ListResponse{
			Index: []*pb.Index{
				&pb.Index{
					Id:   job_.Id(),
					Name: job_.Name(),
				},
			},
		}, nil
	}
}

func (this *index_service) Delete(ctx context.Context, req *pb.IndexId) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.fsindexer.index.Delete>{ req=%v }", req)
	// TODO: Delete an existing index
	return &empty.Empty{}, nil
}

func (this *index_service) Reindex(ctx context.Context, req *pb.IndexId) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.fsindexer.index.Reindex>{ req=%v }", req)
	// TODO: Reindex an existing index
	return &empty.Empty{}, nil
}
