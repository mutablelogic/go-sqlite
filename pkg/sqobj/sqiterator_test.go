package sqobj_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	// Modules
	sqlite3 "github.com/mutablelogic/go-sqlite/pkg/sqlite3"

	// Namespace importst
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
	. "github.com/mutablelogic/go-sqlite/pkg/sqobj"
)

type TestIteratorStructA struct {
	A int    `sqlite:"a,auto"`
	B string `sqlite:"b"`
}

func (t *TestIteratorStructA) String() string {
	return fmt.Sprintf("<TestIteratorStructA a=%d b=%q>", t.A, t.B)
}

func Test_Iterator_001(t *testing.T) {
	conn, err := sqlite3.New()
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	// Set up tracing function
	conn.SetTraceHook(func(conn *sqlite3.Conn, sql string, d time.Duration) {
		if d >= 0 {
			t.Log("EXEC:", sql, "=>", d)
		}
	})

	class, err := RegisterClass(N("test"), TestIteratorStructA{A: 100})
	if err != nil {
		t.Fatal(err)
	}

	// Create
	conn.Do(context.Background(), 0, func(txn SQTransaction) error {
		if err := class.Create(txn, "main"); err != nil {
			t.Error(err)
			return err
		}

		// Return success
		return nil
	})

	// Insert and read ten items
	conn.Do(context.Background(), 0, func(txn SQTransaction) error {
		// Insert ten items
		for i := 0; i < 10; i++ {
			if r, err := class.Insert(txn, TestIteratorStructA{B: fmt.Sprint("T", i+1)}); err != nil {
				t.Error(err)
			} else {
				t.Log(r)
			}
		}

		// Read back ten items
		iter, err := class.Read(txn)
		if err != nil {
			t.Fatal(err)
		}
		for {
			v := iter.Next()
			if v == nil {
				break
			}
			t.Log(v)
		}
		// Return success
		return nil
	})
}
