package sqobj_test

import (
	"fmt"
	"testing"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"

	// Namespace importst
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
	. "github.com/djthorpe/go-sqlite/pkg/sqobj"
)

type TestIteratorStructA struct {
	A int `sqlite:"a,auto"`
	B string
}

func (t *TestIteratorStructA) String() string {
	return fmt.Sprintf("<TestIteratorStructA a=%d b=%q>", t.A, t.B)
}

func Test_Iterator_001(t *testing.T) {
	conn, err := sqlite.New()
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	class, err := RegisterClass(N("test"), TestIteratorStructA{A: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := class.Create(conn, "", SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Fatal(err)
	}

	// Insert ten items
	for i := 0; i < 10; i++ {
		if r, err := class.Insert(conn, TestIteratorStructA{B: fmt.Sprint("T", i+1)}); err != nil {
			t.Error(err)
		} else {
			t.Log(r)
		}
	}

	// Read back ten items
	iter, err := class.Read(conn)
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()
	for {
		v := iter.Next()
		if v == nil {
			break
		}
		fmt.Println(v)
	}
}
