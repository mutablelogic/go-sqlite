package sqlite_test

import (
	"testing"

	sq "github.com/djthorpe/go-sqlite/pkg/sqlite"
)

func Test_Schema_001(t *testing.T) {
	db, err := sq.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema := db.Schemas()
	if schema == nil {
		t.Fatal("Schema is nil")
	}
	if len(schema) != 1 {
		t.Fatal("Schema length not 1")
	}
	if schema[0] != "main" {
		t.Fatal("Schema not 'main'")
	}
}

func Test_Schema_002(t *testing.T) {
	db, err := sq.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	st := db.N("foo").CreateTable(db.N("a").WithType("TEXT"))
	if _, err := db.Exec(st); err != nil {
		t.Fatalf("%q: %v", st.Query(), err)
	}
	tables := db.Tables()
	if tables == nil {
		t.Fatal("tables is nil")
	}
	if len(tables) != 1 {
		t.Fatal("tables length not 1")
	}
	if tables[0] != "foo" {
		t.Fatal("Tables not 'foo'")
	}
}
