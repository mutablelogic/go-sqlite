package sqlite3_test

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/mutablelogic/go-sqlite/sys/sqlite3"
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
	if err := db.ExecEx("SELECT rowid, name FROM file", func(row, col []string) bool {
		if rowid, err := strconv.ParseInt(row[0], 0, 64); err != nil {
			// Abort query
			t.Error(err)
			return true
		} else if blob, err := db.OpenBlob("main", "file", "data", rowid, sqlite3.OpenFlags(sqlite3.SQLITE_OPEN_READWRITE)); err != nil {
			// Abort query
			t.Error(err)
			return true
		} else if data, err := ioutil.ReadFile(filepath.Join(wd, row[1])); err != nil {
			t.Error(err)
			return true
		} else if err := blob.WriteAt(data, 0); err != nil {
			t.Error(err)
			return true
		} else if err := blob.Close(); err != nil {
			t.Error(err)
			return true
		}
		// Success - return 0
		return false
	}); err != nil {
		t.Fatal(err)
	}

	// Retrieve rows, and read from blob files
	if err := db.ExecEx("SELECT rowid, name FROM file", func(row, col []string) bool {
		rowid, err := strconv.ParseInt(row[0], 0, 64)
		if err != nil {
			// Abort query
			t.Error(err)
			return true
		}
		blob, err := db.OpenBlob("main", "file", "data", rowid, sqlite3.OpenFlags(0))
		if err != nil {
			// Abort query
			t.Error(err)
			return true
		}
		defer blob.Close()
		data, err := ioutil.ReadFile(filepath.Join(wd, row[1]))
		if err != nil {
			t.Error(err)
			return true
		}
		data2 := make([]byte, len(data))
		if err := blob.ReadAt(data2, 0); err != nil {
			t.Error(err)
			return true
		}
		if equalsData(data, data2) == false {
			t.Error("Data does not match between file and database")
			return true
		}
		// Success - return 0
		return false
	}); err != nil {
		t.Fatal(err)
	}
}

func Test_Blob_002(t *testing.T) {
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
	if err := db.ExecEx("SELECT rowid, name FROM file", func(row, col []string) bool {
		if rowid, err := strconv.ParseInt(row[0], 0, 64); err != nil {
			// Abort query
			t.Error(err)
			return true
		} else if blob, err := db.OpenBlobEx("main", "file", "data", rowid, sqlite3.OpenFlags(sqlite3.SQLITE_OPEN_READWRITE)); err != nil {
			// Abort query
			t.Error(err)
			return true
		} else if data, err := ioutil.ReadFile(filepath.Join(wd, row[1])); err != nil {
			t.Error(err)
			return true
		} else if _, err := blob.Write(data); err != nil {
			t.Error(err)
			return true
		} else if err := blob.Close(); err != nil {
			t.Error(err)
			return true
		}
		// Success - return 0
		return false
	}); err != nil {
		t.Fatal(err)
	}

	// Retrieve rows, and read from blob files
	if err := db.ExecEx("SELECT rowid, name FROM file", func(row, col []string) bool {
		rowid, err := strconv.ParseInt(row[0], 0, 64)
		if err != nil {
			// Abort query
			t.Error(err)
			return true
		}
		blob, err := db.OpenBlobEx("main", "file", "data", rowid, sqlite3.OpenFlags(0))
		if err != nil {
			// Abort query
			t.Error(err)
			return true
		}
		defer blob.Close()
		data, err := ioutil.ReadFile(filepath.Join(wd, row[1]))
		if err != nil {
			t.Error(err)
			return true
		}
		data2 := make([]byte, len(data))
		if _, err := blob.Read(data2); err != nil {
			t.Error(err)
			return true
		}
		if equalsData(data, data2) == false {
			t.Error("Data does not match between file and database")
			return true
		}
		// Success - return 0
		return false
	}); err != nil {
		t.Fatal(err)
	}
}

func equalsData(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return hex.EncodeToString(a) == hex.EncodeToString(b)
}
