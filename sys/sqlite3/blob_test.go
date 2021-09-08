package sqlite3_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/djthorpe/go-sqlite/sys/sqlite3"
)

func Test_Blob_001(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	db, err := sqlite3.OpenPathEx(filepath.Join(tmpdir, "test.sqlite"), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	if err := db.Exec("CREATE TABLE file (name TEXT PRIMARY KEY,data BLOB)", nil); err != nil {
		t.Fatal(err)
	}

	// Insert zero-blobs
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	files, err := ioutil.ReadDir(wd)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Begin(sqlite3.SQLITE_TXN_DEFAULT); err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if err := db.ExecEx("INSERT INTO file (name,data) VALUES (?,ZEROBLOB(?))", nil, file.Name(), file.Size()); err != nil {
			t.Fatal(err)
		} else if db.Changes() != 1 {
			t.Error("Unexpected return from db.Changes()")
		} else {
			t.Logf("Inserted: %q => %v", file.Name(), db.LastInsertId())
		}
	}
	if err := db.Commit(); err != nil {
		t.Fatal(err)
	}

	// Retrieve rows, and write to blob files
	if err := db.ExecEx("SELECT rowid,name FROM file", func(row, col []string) bool {
		if rowid, err := strconv.ParseInt(row[0], 0, 64); err != nil {
			// Abort query
			t.Error(err)
			return true
		} else if blob, err := db.OpenBlob("main", "file", "data", rowid, sqlite3.OpenFlags(0)); err != nil {
			// Abort query
			t.Error(err)
			return true
		} else if err := blob.Close(); err != nil {
			// Abort query
			t.Error(err)
			return true
		} else {
			t.Logf("%q => %v", row[1], blob)
		}
		// Success - return 0
		return false
	}); err != nil {
		t.Fatal(err)
	}
}
