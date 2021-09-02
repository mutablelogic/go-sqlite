package sqlite_test

import (
	"testing"

	. "github.com/djthorpe/go-sqlite/pkg/lang"
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
	st := N("foo").CreateTable(N("a").WithType("TEXT"))
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
func Test_Schema_003(t *testing.T) {
	db, err := sq.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	modules := db.Modules("")
	if modules == nil {
		t.Error("Modules returned nil")
	}
	for _, module := range modules {
		if modules := db.Modules(module); len(modules) == 0 {
			t.Errorf("Modules with arg %q expected non-empty return", module)
		} else {
			t.Logf("Module(%q) => %q", module, modules)
		}
	}
}

func Test_Schema_004(t *testing.T) {
	db, err := sq.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Create a table
	if _, err := db.Exec(N("foo").CreateTable(N("a").WithType("TEXT").WithPrimary())); err != nil {
		t.Fatal(err)
	}
	// Create an index
	if _, err := db.Exec(N("foo_index").CreateIndex("foo", "a")); err != nil {
		t.Fatal(err)
	}
	// Create a unique index
	if _, err := db.Exec(N("bar_index").CreateIndex("foo", "a").WithUnique()); err != nil {
		t.Fatal(err)
	}
	// Get indexes for table foo
	result := db.Indexes("foo")
	n := 0
	for _, index := range result {
		t.Log(index)
		if index.Auto() {
			continue
		}
		n = n + 1
		switch index.Name() {
		case "foo_index":
			if index.Unique() {
				t.Errorf("Didn't expect unique")
			}
			if index.Table() != "foo" {
				t.Errorf("Unexpected table name %q", index.Table())
			}
			if cols := index.Columns(); len(cols) != 1 || cols[0] != "a" {
				t.Errorf("Unexpected table columns %q", cols)
			}
		case "bar_index":
			if index.Unique() == false {
				t.Errorf("Expected unique")
			}
			if index.Table() != "foo" {
				t.Errorf("Unexpected table name %q", index.Table())
			}
			if cols := index.Columns(); len(cols) != 1 || cols[0] != "a" {
				t.Errorf("Unexpected table columns %q", cols)
			}
		default:
			t.Errorf("Unexpected index name: %q", index.Name())
		}
	}
	if n != 2 {
		t.Errorf("Expected 2 indexes, got %d", n)
	}
}
