package sqlite_test

import (
	"errors"
	"testing"

	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"

	// Imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_Conn_002(t *testing.T) {
	if db, err := sqlite.New(); err != nil {
		t.Fatal(err)
	} else if err := db.Close(); err != nil {
		t.Fatal(err)
	} else {
		t.Log("db=", db)
	}
}

func Test_Conn_003(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if schemas := db.Schemas(); schemas == nil {
		t.Fatal("Unexpected nil from schemas")
	} else if len(schemas) != 1 {
		t.Error("Unexpected number of schemas returned")
	} else if schemas[0] != "main" {
		t.Errorf("Unexpected schema name, %q", schemas[0])
	}
}

func Test_Conn_004(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(Q("CREATE TABLE foo (id INTEGER PRIMARY KEY, name TEXT)")); err != nil {
		t.Fatal(err)
	}
	if r, err := db.Exec(Q("INSERT INTO foo (id, name) VALUES (1, 'bar')")); err != nil {
		t.Error(err)
	} else if r.LastInsertId != 1 {
		t.Errorf("Unexpected LastInsertId, %d", r.LastInsertId)
	} else if r.RowsAffected != 1 {
		t.Errorf("Unexpected RowsAffected, %d", r.RowsAffected)
	}
	if r, err := db.Exec(N("foo").Insert()); err != nil {
		t.Error(err)
	} else if r.LastInsertId != 2 {
		t.Errorf("Unexpected LastInsertId, %d", r.LastInsertId)
	} else if r.RowsAffected != 1 {
		t.Errorf("Unexpected RowsAffected, %d", r.RowsAffected)
	}
	if r, err := db.Exec(N("foo").Insert("id", "name"), 10, "name"); err != nil {
		t.Error(err)
	} else if r.LastInsertId != 10 {
		t.Errorf("Unexpected LastInsertId, %d", r.LastInsertId)
	} else if r.RowsAffected != 1 {
		t.Errorf("Unexpected RowsAffected, %d", r.RowsAffected)
	}
	rows, err := db.Query(Q("SELECT * FROM foo"))
	if err != nil {
		t.Error(err)
	}
	for {
		row := rows.NextMap()
		if row == nil {
			break
		}
		t.Logf("%+v", row)
	}
}
func Test_Conn_005(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Do(func(txn SQTransaction) error {
		_, err := txn.Exec(Q("CREATE TABLE foo (id INTEGER PRIMARY KEY, name TEXT)"))
		return err
	}); err != nil {
		t.Fatal(err)
	}

	if err := db.Do(func(txn SQTransaction) error {
		for i := 0; i < 5; i++ {
			if _, err := txn.Exec(Q("INSERT INTO foo DEFAULT VALUES")); err != nil {
				return err
			}
		}
		return ErrNotImplemented
	}); err != nil && errors.Is(err, ErrNotImplemented) == false {
		t.Fatal(err)
	}
}
