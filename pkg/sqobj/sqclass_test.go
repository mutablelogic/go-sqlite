package sqobj_test

import (
	"testing"

	// Modules
	. "github.com/djthorpe/go-sqlite"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/sqobj"
)

type TestClassStructA struct {
	A int `sqlite:"a,auto"`
}

func Test_Class_000(t *testing.T) {
	class, err := NewClass("test", "", TestClassStructA{A: 100})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(class)
}

func Test_Class_001(t *testing.T) {
	conn, err := sqlite.New()
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	class, err := NewClass("test", "", TestClassStructA{A: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := class.Create(conn, SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Error(err)
	} else if err := class.Create(conn, SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Error(err)
	} else {
		t.Log(conn.Tables())
	}
}
