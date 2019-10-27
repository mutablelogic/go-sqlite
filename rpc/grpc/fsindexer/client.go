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
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"
	sq "github.com/djthorpe/sqlite"
	empty "github.com/golang/protobuf/ptypes/empty"

	// Protocol buffers
	pb "github.com/djthorpe/sqlite/rpc/protobuf/fsindexer"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type client struct {
	conn gopi.RPCClientConn
}

type IndexerClient struct {
	pb.IndexerClient
	client
}

type QueryClient struct {
	pb.QueryClient
	client
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewIndexerClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &IndexerClient{pb.NewIndexerClient(conn.(grpc.GRPCClientConn).GRPCConn()), client{conn}}
}

func NewQueryClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &QueryClient{pb.NewQueryClient(conn.(grpc.GRPCClientConn).GRPCConn()), client{conn}}
}

func (this *client) NewContext(timeout time.Duration) context.Context {
	if timeout == 0 {
		timeout = this.conn.Timeout()
	}
	if timeout == 0 {
		return context.Background()
	} else {
		ctx, _ := context.WithTimeout(context.Background(), timeout)
		return ctx
	}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *client) Conn() gopi.RPCClientConn {
	return this.conn
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *IndexerClient) String() string {
	return fmt.Sprintf("<fsindexer.client.indexer>{ %v }", this.Conn())
}

func (this *QueryClient) String() string {
	return fmt.Sprintf("<fsindexer.client.query>{ %v }", this.Conn())
}

////////////////////////////////////////////////////////////////////////////////
// CALLS

func (this *IndexerClient) Ping() error {
	this.conn.Lock()
	defer this.conn.Unlock()
	if _, err := this.IndexerClient.Ping(this.NewContext(0), &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *QueryClient) Ping() error {
	this.conn.Lock()
	defer this.conn.Unlock()
	if _, err := this.QueryClient.Ping(this.NewContext(0), &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *IndexerClient) AddIndex(path string, watch bool) (sq.FSIndex, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Perform index command
	if reply, err := this.IndexerClient.AddIndex(this.NewContext(0), &pb.IndexRequest{
		Path:  path,
		Watch: watch,
	}); err != nil {
		return nil, err
	} else {
		return &fsindex_proto{reply}, nil
	}
}

func (this *IndexerClient) DeleteIndex(index int64) error {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Perform index command
	if _, err := this.IndexerClient.DeleteIndex(this.NewContext(0), &pb.IndexId{
		Id: index,
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *IndexerClient) List() ([]sq.FSIndex, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Obtain list of indexes
	if reply, err := this.IndexerClient.List(this.NewContext(0), &empty.Empty{}); err != nil {
		return nil, err
	} else {
		indexes := make([]sq.FSIndex, len(reply.Index))
		for i := range reply.Index {
			indexes[i] = &fsindex_proto{reply.Index[i]}
		}
		return indexes, nil
	}

}

func (this *QueryClient) List() ([]sq.FSIndex, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Obtain list of indexes
	if reply, err := this.QueryClient.List(this.NewContext(0), &empty.Empty{}); err != nil {
		return nil, err
	} else {
		indexes := make([]sq.FSIndex, len(reply.Index))
		for i := range reply.Index {
			indexes[i] = &fsindex_proto{reply.Index[i]}
		}
		return indexes, nil
	}
}

func (this *QueryClient) Query(limit uint64) (sq.FSQueryResponse, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Obtain list of indexes
	if reply, err := this.QueryClient.Query(this.NewContext(0), &pb.QueryRequest{
		Limit: limit,
	}); err != nil {
		return nil, err
	} else {
		return NewQueryResponse(reply), nil
	}
}
