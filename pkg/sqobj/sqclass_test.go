package sqobj_test

import (
	"context"
	"testing"
	"time"

	// Package imports
	"github.com/djthorpe/go-sqlite"
	"github.com/djthorpe/go-sqlite/pkg/sqlite3"

	// Namespace imports
	. "github.com/djthorpe/go-sqlite/pkg/lang"
	. "github.com/djthorpe/go-sqlite/pkg/sqobj"
)

type TestClassStructA struct {
	A int `sqlite:"a,auto"`
}

func Test_Class_000(t *testing.T) {
	class := MustRegisterClass(N("test"), TestClassStructA{A: 100})
	t.Log(class)
}

func Test_Class_001(t *testing.T) {
	conn, err := sqlite3.OpenPath(":memory:", sqlite.SQLITE_OPEN_OVERWRITE)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	class, err := RegisterClass(N("test"), TestClassStructA{A: 100})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(class)
	}
	if err := class.Create(context.Background(), conn, ""); err != nil {
		t.Error(err)
	} else if err := class.Create(context.Background(), conn, ""); err != nil {
		t.Error(err)
	} else {
		t.Log(conn.Tables(""))
	}
}

type TestClassStructB struct {
	Index string `sqlite:"index,primary" json:"index"`
	Path  string `sqlite:"path,primary" json:"path"`
	Name  string `sqlite:"name,primary" json:"name"`
}

type TestClassStructC struct {
	Index string `sqlite:"index,primary,foreign"`
	Path  string `sqlite:"path,primary,foreign"`
	Name  string `sqlite:"name,primary,foreign"`
}

func Test_Class_002(t *testing.T) {
	cFile := MustRegisterClass(N("file"), &TestClassStructB{})
	cFileMark := MustRegisterClass(N("mark"), &TestClassStructC{}).ForeignKey(cFile)

	db, err := sqlite3.New(sqlite.SQLITE_OPEN_OVERWRITE)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set up tracing function
	db.SetTraceHook(func(sql string, d time.Duration) {
		if d >= 0 {
			t.Log("EXEC:", sql, "=>", d)
		}
	})

	// Create twice
	if err := cFile.Create(context.Background(), db, "main"); err != nil {
		t.Error(err)
	} else if err := cFileMark.Create(context.Background(), db, "main"); err != nil {
		t.Error(err)
	} else {
		t.Log("First pass:", db.Tables("main"))
	}

	if err := cFile.Create(context.Background(), db, "main"); err != nil {
		t.Error(err)
	} else if err := cFileMark.Create(context.Background(), db, "main"); err != nil {
		t.Error(err)
	} else {
		t.Log("Second pass:", db.Tables("main"))
	}
}

func Test_Class_003(t *testing.T) {
	cFile := MustRegisterClass(N("file"), &TestClassStructB{})

	db, err := sqlite3.New(sqlite.SQLITE_OPEN_OVERWRITE)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set up tracing function
	db.SetTraceHook(func(sql string, d time.Duration) {
		if d >= 0 {
			t.Log("EXEC:", sql, "=>", d)
		}
	})

	// Create
	if err := cFile.Create(context.Background(), db, "main"); err != nil {
		t.Error(err)
	} else {
		t.Log("First pass:", db.Tables("main"))
	}

	// Insert two rows
	db.Do(context.Background(), 0, func(txn sqlite.SQTransaction) error {
		if rows, err := cFile.Insert(txn, TestClassStructB{"A", "B", "C"}, TestClassStructB{"D", "E", "F"}); err != nil {
			t.Error(err)
			return err
		} else {
			t.Log("rows=", rows)
		}
		// Return success
		return nil
	})
}

type TestClassStructD struct {
	Key int `sqlite:"key,auto"`
}

func Test_Class_004(t *testing.T) {
	cKey := MustRegisterClass(N("key"), TestClassStructD{})

	db, err := sqlite3.New(sqlite.SQLITE_OPEN_OVERWRITE)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set up tracing function
	db.SetTraceHook(func(sql string, d time.Duration) {
		if d >= 0 {
			t.Log("EXEC:", sql, "=>", d)
		}
	})

	// Create
	if err := cKey.Create(context.Background(), db, "main"); err != nil {
		t.Error(err)
	}

	// Insert two rows - should NULL the auto increment value
	db.Do(context.Background(), 0, func(txn sqlite.SQTransaction) error {
		if rows, err := cKey.Insert(txn, TestClassStructD{1}, TestClassStructD{1}); err != nil {
			t.Error(err)
			return err
		} else {
			t.Log("rows=", rows)
		}
		// Return success
		return nil
	})
}
