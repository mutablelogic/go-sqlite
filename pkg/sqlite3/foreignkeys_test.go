package sqlite3_test

import (
	"context"
	"testing"

	// Module imports

	// Namespace Imports
	. "github.com/djthorpe/go-sqlite/pkg/sqlite3"
)

func Test_ForeignKeys_001(t *testing.T) {
	errs, cancel := catchErrors(t)
	defer cancel()

	// Create the pool
	pool, err := OpenPool(PoolConfig{
		Schemas: map[string]string{"main": ":memory:"},
		Trace:   true,
	}, errs)
	if err != nil {
		t.Error(err)
	}
	defer pool.Close()

	// Get conn
	conn := pool.Get(context.Background())
	if conn == nil {
		t.Fatal("conn is nil")
	}
	defer pool.Put(conn)

	if err := conn.(*Conn).SetForeignKeyConstraints(true); err != nil {
		t.Error(err)
	} else if v, err := conn.(*Conn).ForeignKeyConstraints(); err != nil {
		t.Error(err)
	} else if v != true {
		t.Error("Unexpected response from ForeignKeyConstraints")
	}

	if err := conn.(*Conn).SetForeignKeyConstraints(false); err != nil {
		t.Error(err)
	} else if v, err := conn.(*Conn).ForeignKeyConstraints(); err != nil {
		t.Error(err)
	} else if v != false {
		t.Error("Unexpected response from ForeignKeyConstraints")
	}
}
