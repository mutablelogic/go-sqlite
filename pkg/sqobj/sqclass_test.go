package sqobj_test

import (
	"context"
	"testing"
	"time"

	// Package imports
	"github.com/djthorpe/go-sqlite"
	"github.com/djthorpe/go-sqlite/pkg/sqlite3"

	// Namespace imports
	. "github.com/djthorpe/go-sqlite"
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

	conn.Do(context.Background(), 0, func(txn SQTransaction) error {
		if err := class.Create(txn, ""); err != nil {
			t.Error(err)
			return err
		} else if err := class.Create(txn, ""); err != nil {
			t.Error(err)
			return err
		} else {
			t.Log(conn.Tables(""))
			return nil
		}
	})
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

	db.Do(context.Background(), 0, func(txn SQTransaction) error {
		// Create first time
		if err := cFile.Create(txn, "main"); err != nil {
			t.Error(err)
		} else if err := cFileMark.Create(txn, "main"); err != nil {
			t.Error(err)
		} else {
			t.Log("First pass:", db.Tables("main"))
		}

		// Create second time
		if err := cFile.Create(txn, "main"); err != nil {
			t.Error(err)
		} else if err := cFileMark.Create(txn, "main"); err != nil {
			t.Error(err)
		} else {
			t.Log("Second pass:", db.Tables("main"))
		}

		// Return success
		return nil
	})
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

	db.Do(context.Background(), 0, func(txn SQTransaction) error {
		// Create first time
		if err := cFile.Create(txn, "main"); err != nil {
			t.Error(err)
		} else {
			t.Log("First pass:", db.Tables("main"))
		}

		// Create second time
		if err := cFile.Create(txn, "main"); err != nil {
			t.Error(err)
		} else {
			t.Log("Second pass:", db.Tables("main"))
		}

		// Return success
		return nil
	})

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
	db.Do(context.Background(), 0, func(txn SQTransaction) error {
		if err := cKey.Create(txn, "main"); err != nil {
			t.Error(err)
			return err
		}

		// Return success
		return nil
	})

	// Insert two rows - should NULL the auto increment value
	db.Do(context.Background(), 0, func(txn sqlite.SQTransaction) error {
		if rows, err := cKey.Insert(txn, TestClassStructD{}, TestClassStructD{}); err != nil {
			t.Error(err)
			return err
		} else if n, err := cKey.DeleteRows(txn, rows); err != nil {
			t.Error(err)
			return err
		} else if len(rows) != n {
			t.Error("Expected", n, "rows deleted, got", len(rows))
		} else {
			t.Log("rows=", rows, " deleted=", n)
		}
		// Return success
		return nil
	})
}

func Test_Class_005(t *testing.T) {
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
	db.Do(context.Background(), 0, func(txn SQTransaction) error {
		if err := cKey.Create(txn, "main"); err != nil {
			t.Error(err)
			return err
		}

		// Return success
		return nil
	})

	// Rows
	r := []interface{}{
		TestClassStructD{1}, TestClassStructD{2},
	}

	// Insert two rows - should NULL the auto increment value
	db.Do(context.Background(), 0, func(txn sqlite.SQTransaction) error {
		if rows, err := cKey.Insert(txn, r...); err != nil {
			t.Error(err)
			return err
		} else if n, err := cKey.DeleteKeys(txn, r...); err != nil {
			t.Error(err)
			return err
		} else if len(rows) != n {
			t.Error("Expected", n, "rows deleted, got", len(rows))
		} else {
			t.Log("rows=", rows, " deleted=", n)
		}
		// Return success
		return nil
	})
}

type TestClassStructE struct {
	KeyA  int `sqlite:"key_a,primary"`
	KeyB  int `sqlite:"key_b,primary"`
	Value string
}

func Test_Class_006(t *testing.T) {
	cKey := MustRegisterClass(N("key"), TestClassStructE{})

	db, err := sqlite3.New(sqlite.SQLITE_OPEN_OVERWRITE)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set up tracing function
	db.SetTraceHook(func(sql string, d time.Duration) {
		if d >= 0 {
			t.Log("EXEC:", t.Name(), sql, "=>", d)
		}
	})

	// Rows
	r := []interface{}{
		&TestClassStructE{0, 0, "Row 1"}, &TestClassStructE{1, 1, "Row 2"}, &TestClassStructE{2, 2, "Row 3"},
	}

	db.Do(context.Background(), 0, func(txn SQTransaction) error {
		if err := cKey.Create(txn, "main"); err != nil {
			t.Error(err)
			return err
		}
		if _, err := cKey.Insert(txn, r...); err != nil {
			t.Error(err)
			return err
		}

		// Update values
		for _, r := range r {
			r.(*TestClassStructE).Value = "Updated " + r.(*TestClassStructE).Value
		}

		if n, err := cKey.UpdateKeys(txn, r...); err != nil {
			t.Error(err)
			return err
		} else if n != len(r) {
			t.Error("Expected", len(r), "rows updated, got", n)
		} else {
			t.Log("Updated =>", n, "rows affected")
		}
		iter, err := cKey.Read(txn)
		if err != nil {
			t.Error(err)
			return err
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

func Test_Class_007(t *testing.T) {
	cKey := MustRegisterClass(N("key"), TestClassStructE{})

	db, err := sqlite3.New(sqlite.SQLITE_OPEN_OVERWRITE)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// Set up tracing function
	db.SetTraceHook(func(sql string, d time.Duration) {
		if d >= 0 {
			t.Log("EXEC:", t.Name(), sql, "=>", d)
		}
	})

	// Rows
	r := []interface{}{
		&TestClassStructE{0, 0, "Row 1"}, &TestClassStructE{1, 1, "Row 2"}, &TestClassStructE{2, 2, "Row 3"},
	}

	db.Do(context.Background(), 0, func(txn SQTransaction) error {
		if err := cKey.Create(txn, "main"); err != nil {
			t.Error(err)
			return err
		}

		// Upsert values - does an insert
		if r, err := cKey.UpsertKeys(txn, r...); err != nil {
			t.Error(err)
			return err
		} else {
			t.Log("insert => ", r)
		}

		// Update values - does not do an update
		if r, err := cKey.UpsertKeys(txn, r...); err != nil {
			t.Error(err)
			return err
		} else {
			t.Log("not update => ", r)
		}

		// Update the second object
		r[1].(*TestClassStructE).Value = "Updated " + r[1].(*TestClassStructE).Value

		// Update values - only update the second one
		if r, err := cKey.UpsertKeys(txn, r...); err != nil {
			t.Error(err)
			return err
		} else {
			t.Log("update 2nd => ", r)
		}

		// Update values
		for _, r := range r {
			r.(*TestClassStructE).Value = "Updated " + r.(*TestClassStructE).Value
		}
		// Add another
		r = append(r, &TestClassStructE{3, 3, "Row 4"})

		// Update values
		if r, err := cKey.UpsertKeys(txn, r...); err != nil {
			t.Error(err)
			return err
		} else {
			t.Log("update 1st and 3rd and insert 4th => ", r)
		}

		// Update values - does not do an update
		if r, err := cKey.UpsertKeys(txn, r...); err != nil {
			t.Error(err)
			return err
		} else {
			t.Log("not update => ", r)
		}

		// Return success
		return nil
	})
}
