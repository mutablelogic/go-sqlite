package sqlite3_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	// Namespace Imports
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
	. "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
)

func Test_Schema_001(t *testing.T) {
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

	// Get schemas
	schemas := conn.Schemas()
	if schemas == nil {
		t.Error("Unexpected return from schemas")
	} else if len(schemas) != 1 {
		t.Error("Unexpected return from schemas")
	} else if schemas[0] != "main" {
		t.Error("Unexpected return from schemas")
	}
}

func Test_Schema_002(t *testing.T) {
	// Create error channel
	errs, cancel := handleErrors(t)

	// Make folder of temp files
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Make configuration
	cfg := NewConfig().WithTrace(func(sql string, d time.Duration) {
		if d > 0 {
			t.Log(sql, "=>", d)
		}
	}).
		WithSchema("main", filepath.Join(tmpdir, "main.sqlite")).
		WithSchema("test", filepath.Join(tmpdir, "test.sqlite"))

	// Create pool
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

	// Make schema request
	schemas := conn.Schemas()

	if schemas == nil {
		t.Errorf("Unexpected return from schemas: %q", schemas)
	} else if len(schemas) != 2 {
		t.Errorf("Unexpected return from schemas: %q", schemas)
	} else {
		t.Logf("schemas: %q", schemas)
	}
}

func Test_Schema_003(t *testing.T) {
	// Create error channel
	errs, cancel := handleErrors(t)

	// Make folder of temp files
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Make configuration
	cfg := NewConfig().WithTrace(func(sql string, d time.Duration) {
		if d > 0 {
			t.Log(sql, "=>", d)
		}
	}).
		WithSchema("main", filepath.Join(tmpdir, "main.sqlite")).
		WithSchema("test", filepath.Join(tmpdir, "test.sqlite"))

	// Create pool
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

	// Create table_a and table_b in main schema
	if conn == nil {
		t.Fatal("Unexpected nil connection")
	}
	if err := conn.Exec(N("table_a").CreateTable(N("a").WithType("TEXT")), nil); err != nil {
		t.Error(err)
	}
	if err := conn.Exec(N("table_b").CreateTable(N("a").WithType("TEXT")), nil); err != nil {
		t.Error(err)
	}

	// Obtain the tables
	tables := conn.Tables("main")
	if tables == nil {
		t.Errorf("Unexpected return from tables: %q", tables)
	} else if len(tables) != 2 {
		t.Errorf("Unexpected return from tables: %q", tables)
	} else {
		t.Logf("tables: %q", tables)
	}

	// Create table_a and table_b as temporary in main schema
	if err := conn.Exec(N("table_a").CreateTable(N("a").WithType("TEXT")).WithTemporary(), nil); err != nil {
		t.Error(err)
	}
	if err := conn.Exec(N("table_b").CreateTable(N("a").WithType("TEXT")).WithTemporary(), nil); err != nil {
		t.Error(err)
	}

	// Obtain the tables
	tables = conn.Tables("temp")
	if tables == nil {
		t.Errorf("Unexpected return from tables: %q", tables)
	} else if len(tables) != 2 {
		t.Errorf("Unexpected return from tables: %q", tables)
	} else {
		t.Logf("tables: %q", tables)
	}
}

func Test_Schema_004(t *testing.T) {
	// Create error channel
	errs, cancel := handleErrors(t)

	// Make folder of temp files
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Make configuration
	cfg := NewConfig().WithTrace(func(sql string, d time.Duration) {
		if d > 0 {
			t.Log(sql, "=>", d)
		}
	}).
		WithSchema("main", filepath.Join(tmpdir, "main.sqlite")).
		WithSchema("test", filepath.Join(tmpdir, "test.sqlite"))

	// Create pool
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

	// Get full module list
	modules := conn.Modules()
	if modules == nil {
		t.Errorf("Unexpected nil return from modules: %q", modules)
	} else {
		t.Logf("modules: %q", modules)
	}

	// Expect one or more modules for return of prefix version
	for _, module := range modules {
		if v := conn.Modules(module); v == nil {
			t.Errorf("Unexpected nil return from modules: %q", v)
		} else if len(v) == 0 {
			t.Errorf("Unexpected return from modules: %q", v)
		} else {
			t.Logf("module(%q) => %q", module, v)
		}
	}
}

func Test_Schema_006(t *testing.T) {
	// Create error channel
	errs, cancel := handleErrors(t)

	// Make configuration
	cfg := NewConfig().WithTrace(func(sql string, d time.Duration) {
		if d > 0 {
			t.Log(sql, "=>", d)
		}
	})

	// Create pool
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

	// Create a table
	if err := conn.Exec(N("table_a").CreateTable(
		C("a").WithType("INTEGER").WithAutoIncrement(),
		C("b").NotNull(),
		C("c").WithType("TIMESTAMP").WithDefault(0),
	), nil); err != nil {
		t.Error(err)
	}

	// Obtain the columns
	columns := conn.ColumnsForTable("main", "table_a")
	if columns == nil {
		t.Errorf("Unexpected return from columns: %q", columns)
	} else if len(columns) != 3 {
		t.Errorf("Unexpected return from columns: %q", columns)
	} else {
		t.Logf("columns: %q", columns)
	}
}

func Test_Schema_007(t *testing.T) {
	// Create error channel
	errs, cancel := handleErrors(t)

	// Make folder of temp files
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Make configuration
	cfg := NewConfig().WithTrace(func(sql string, d time.Duration) {
		if d > 0 {
			t.Log(sql, "=>", d)
		}
	}).
		WithSchema("main", filepath.Join(tmpdir, "main.sqlite")).
		WithSchema("test", filepath.Join(tmpdir, "test.sqlite"))

	// Create pool
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

	// Create a table
	if err := conn.Exec(N("table_a").CreateTable(
		C("a").WithType("INTEGER").WithAutoIncrement(),
		C("b").NotNull(),
		C("c").WithType("TIMESTAMP"),
	), nil); err != nil {
		t.Error(err)
	}

	// Create indexes
	if err := conn.Exec(N("index_a").CreateIndex("table_a", "a", "b"), nil); err != nil {
		t.Error(err)
	}

	// Obtain the indexes
	indexes := conn.IndexesForTable("main", "table_a")
	if indexes == nil {
		t.Errorf("Unexpected return from indexes: %q", indexes)
	} else if len(indexes) != 1 {
		t.Errorf("Unexpected return from indexes: %q", indexes)
	} else {
		t.Logf("indexes: %q", indexes)
	}
}
