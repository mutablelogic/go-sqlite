package sqlite3_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/djthorpe/go-sqlite/sys/sqlite3"
)

func Test_SQLite_001(t *testing.T) {
	version, number, id := sqlite3.Version()
	t.Logf("Version: %v", version)
	t.Logf("Number: %v", number)
	t.Logf("ID: %v", id)
}

func Test_SQLite_002(t *testing.T) {
	db, err := sqlite3.OpenPath(":memory:", sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	} else if err := db.Close(); err != nil {
		t.Error(err)
	} else {
		t.Log(db)
	}
}
func Test_SQLite_003(t *testing.T) {
	db, err := sqlite3.OpenPathEx(":memory:", sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	} else if err := db.SetBusyTimeout(5 * time.Second); err != nil {
		t.Error(err)
	} else if err := db.Close(); err != nil {
		t.Error(err)
	}
}
func Test_SQLite_004(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	db, err := sqlite3.OpenPath(filepath.Join(tmpdir, "test.sqlite"), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	for i := sqlite3.SQLITE_LIMIT_MIN; i <= sqlite3.SQLITE_LIMIT_MAX; i++ {
		value := db.GetLimit(i)
		if prev := db.SetLimit(i, value); prev != value {
			t.Errorf("Unexpected return from SetLimit(%v, %v)", i, value)
		} else {
			t.Logf("%v => %v", i, value)
		}
	}
}

func Test_SQLite_005(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	db, err := sqlite3.OpenPath(filepath.Join(tmpdir, "test.sqlite"), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	if filename := db.Filename(""); filename == "" {
		t.Error("Unexpected return from Filename")
	} else {
		t.Log("Filename=", filename)
	}

	if statement, extra, err := db.Prepare("SELECT NULL; SELECT NULL"); err != nil {
		t.Error(err)
	} else if err := statement.Finalize(); err != nil {
		t.Error(err)
	} else {
		t.Log("Statement=", statement)
		t.Log("Extra=", extra)
	}
}

func Test_SQLite_006(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	db, err := sqlite3.OpenPath(filepath.Join(tmpdir, "test.sqlite"), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	create, _, err := db.Prepare("CREATE TABLE test (a TEXT,b TEXT)")
	if err != nil {
		t.Error(err)
	}
	defer create.Finalize()
	for {
		st := create.Step()
		if st == sqlite3.SQLITE_DONE {
			break
		}
		fmt.Println(st)
	}

	statement, _, err := db.Prepare("SELECT * FROM test WHERE a=:a AND b=:b")
	if err != nil {
		t.Error(err)
	}
	defer statement.Finalize()

	for i := 0; i < statement.NumParams(); i++ {
		name := statement.ParamName(i + 1)
		t.Log("Param ", i+1, "=", name, " => ", statement.ParamIndex(name))
	}

	for i := 0; i < statement.ColumnCount(); i++ {
		t.Log("Column ", i)
		t.Log(" Name:", statement.ColumnName(i))
		t.Log(" Database:", statement.ColumnDatabaseName(i))
		t.Log(" Type:", statement.ColumnType(i))
		t.Log(" DeclType:", statement.ColumnDeclType(i))
		t.Log(" Origin:", statement.ColumnOriginName(i))
	}
}

func Test_SQLite_007(t *testing.T) {
	if sqlite3.IsComplete("SELECT") == true {
		t.Error("Unexpected response")
	}
	if sqlite3.IsComplete("SELECT NULL;") == false {
		t.Error("Unexpected response")
	}
}

func Test_SQLite_008(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	db, err := sqlite3.OpenPath(filepath.Join(tmpdir, "test.sqlite"), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	for i := sqlite3.SQLITE_DBSTATUS_MIN; i <= sqlite3.SQLITE_DBSTATUS_MAX; i++ {
		if cur, max, err := db.GetStatus(i); err != nil {
			t.Error(err)
		} else {
			t.Log("  ", i, " cur=", cur, " max=", max)
		}
	}

	cur, max := sqlite3.GetMemoryUsed()
	t.Logf("   Memory used cur=%v max=%v", cur, max)
}

func Test_SQLite_009(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	db, err := sqlite3.OpenPath(filepath.Join(tmpdir, "test.sqlite"), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	create, _, err := db.Prepare("CREATE TABLE test (a TEXT,b TEXT)")
	if err != nil {
		t.Error(err)
	}
	defer create.Finalize()
	for {
		st := create.Step()
		if st == sqlite3.SQLITE_DONE {
			break
		}
		fmt.Println(st)
	}

	statement, _, err := db.Prepare("SELECT * FROM test WHERE a=:a AND b=:b")
	if err != nil {
		t.Error(err)
	}
	defer statement.Finalize()

	t.Log("SQL=", statement.SQL())
	t.Log("ExpandedSQL=", statement.ExpandedSQL())

	var s *sqlite3.Statement
	for {
		s = db.NextStatement(s)
		if s == nil {
			break
		} else {
			t.Log(s)
		}
	}
}

func Test_SQLite_010(t *testing.T) {
	for i := 0; i < sqlite3.KeywordCount(); i++ {
		name := sqlite3.KeywordName(i)
		t.Log("Keyword ", i, "=>", name, "=>", sqlite3.KeywordCheck(name))
	}
}
