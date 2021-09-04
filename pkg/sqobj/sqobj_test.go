package sqobj_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/sqobj"
)

type doc struct {
	A int    `sqlite:"a,autoincrement"`
	B string `sqlite:"b"`
}

func Test_Objects_000(t *testing.T) {
	if db, err := New(); err != nil {
		t.Fatal(err)
	} else {
		defer db.Close()
		t.Log(db)
	}
}

func Test_Objects_001(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Register the document type, create schema twice (second time will also drop)
	class, err := db.Register("doc", doc{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Create(class, SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(class, SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Fatal(err)
	}
}

func Test_Objects_002(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Register the document type, create schema twice (second time will also drop)
	class, err := db.Register("doc", doc{})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(class)
	}
	if err := db.Create(class, SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Fatal(err)
	} else if err := db.Create(class, SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Fatal(err)
	} else {
		t.Log("CREATE:", class.Get(SQKeyCreate))
		t.Log("WRITE:", class.Get(SQKeyWrite))
		t.Log("READ:", class.Get(SQKeyRead))
	}

	// Write three values
	for i := 0; i < 3; i++ {
		if r, err := db.Write(doc{B: fmt.Sprintf("Test %d", i)}); err != nil {
			t.Fatal(err)
		} else if r[0].RowsAffected != 1 {
			t.Error("Unexpected rows affected:", r[0].RowsAffected)
		} else {
			t.Log("WRITE:", i, " lastinsertid=", r[0].LastInsertId)
		}
	}

	// Read three values
	rs, err := db.Read(class)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	for {
		obj, ok := rs.Next().(*doc)
		if !ok {
			break
		}
		fmt.Println(obj)
	}
}
