package sqobj_test

import (
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
	}

	// Write three documents - separate documents with autoincrementing primary key
	if result, err := db.Write(doc{}, doc{}, doc{}); err != nil {
		t.Error(err)
	} else {
		t.Log(result)
	}

	// Write a single document then update it
	v := doc{100, "b"}
	if result, err := db.Write(v); err != nil {
		t.Error(err)
	} else {
		t.Log(result)
	}
	if result, err := db.Write(v); err != nil {
		t.Error(err)
	} else {
		t.Log(result)
	}
}
