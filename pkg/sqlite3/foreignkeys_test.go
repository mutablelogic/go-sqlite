package sqlite3_test

import (
	"testing"
	"time"

	// Module imports

	// Namespace Imports
	. "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
)

func Test_ForeignKeys_001(t *testing.T) {
	errs, cancel := handleErrors(t)
	cfg := NewConfig().WithTrace(func(sql string, d time.Duration) {
		if d > 0 {
			t.Log(sql, "=>", d)
		}
	})
	pool, err := OpenPool(cfg, errs)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(pool)
	}
	defer pool.Close()
	defer cancel()

	// Get connection
	conn := pool.Get()
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
