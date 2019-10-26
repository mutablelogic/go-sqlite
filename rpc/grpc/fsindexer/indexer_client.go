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

type Client struct {
	pb.IndexerClient
	conn gopi.RPCClientConn
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewIndexerClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &Client{pb.NewIndexerClient(conn.(grpc.GRPCClientConn).GRPCConn()), conn}
}

func (this *Client) NewContext(timeout time.Duration) context.Context {
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

func (this *Client) Conn() gopi.RPCClientConn {
	return this.conn
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Client) String() string {
	return fmt.Sprintf("<fsindexer.client.indexer>{ %v }", this.Conn())
}

////////////////////////////////////////////////////////////////////////////////
// CALLS

func (this *Client) Ping() error {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Perform ping
	if _, err := this.IndexerClient.Ping(this.NewContext(0), &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *Client) AddIndex(path string, watch bool) (sq.FSIndex, error) {
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

func (this *Client) DeleteIndex(index int64) error {
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

func (this *Client) List() ([]sq.FSIndex, error) {
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