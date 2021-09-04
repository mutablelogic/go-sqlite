package sqobj_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/sqobj"
)

// document with an autoincrement field
type doc struct {
	A int    `sqlite:"a,autoincrement"`
	B string `sqlite:"b"`
}

func (d *doc) String() string {
	return fmt.Sprintf("<doc %d %q>", d.A, d.B)
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

	// Write 10 values
	for i := 0; i < 10; i++ {
		if r, err := db.Write(doc{B: fmt.Sprintf("%d", i+1)}); err != nil {
			t.Fatal(err)
		} else if r[0].RowsAffected != 1 {
			t.Error("Unexpected rows affected:", r[0].RowsAffected)
		} else {
			t.Log("WRITE:", " lastinsertid=", r[0].LastInsertId)
		}
	}

	// Read all values and re-insert
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
		if r, err := db.Write(obj); err != nil {
			t.Error(err)
		} else {
			t.Log(obj, "=>", r[0])
		}
		obj.B = "new " + obj.B
		if r, err := db.Write(obj); err != nil {
			t.Error(err)
		} else {
			t.Log(obj, "=>", r[0])
		}
	}
}

// column with a unique constraint
type doc2 struct {
	A int    `sqlite:"a,autoincrement"`
	B int    `sqlite:"b,unique"`
	C string `sqlite:"c"`
}

func Test_Objects_003(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Register the document type, create schema twice (second time will also drop)
	class, err := db.Register("doc", doc2{})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(class)
	}
	if err := db.Create(class, SQLITE_FLAG_DELETEIFEXISTS|SQLITE_FLAG_UPDATEONINSERT); err != nil {
		t.Fatal(err)
	}
	t.Log("CREATE:", class.Get(SQKeyCreate))
	t.Log("WRITE:", class.Get(SQKeyWrite))
	t.Log("READ:", class.Get(SQKeyRead))

	// Write three objects
	a := doc2{B: 1, C: "one"}
	b := doc2{B: 2, C: "two"}
	c := doc2{B: 2, C: "three"}
	if r, err := db.Write(a, b, c); err != nil {
		t.Error(err)
	} else {
		t.Log(a, b, c, "=>", r)
	}
}
