package sqlite3_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/djthorpe/go-sqlite/sys/sqlite3"
)

func Test_Backup_001(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Open source
	src, err := sqlite3.OpenPathEx(filepath.Join(tmpdir, "src.sqlite"), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	}
	defer src.Close()

	// Open destination
	dest, err := sqlite3.OpenPathEx(filepath.Join(tmpdir, "dest.sqlite"), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	}
	defer dest.Close()

	// Add documents to src
	if err := src.Exec("CREATE TABLE test (a INTEGER PRIMARY KEY)", nil); err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= 9999; i++ {
		if err := src.Exec("INSERT INTO TEST DEFAULT VALUES", nil); err != nil {
			t.Fatal(err)
		}
	}

	// Backup to dest
	backup, err := src.OpenBackup(dest.Conn, "", "")
	if err != nil {
		t.Error(err)
	}
	defer backup.Finish()

	for {
		if err := backup.Step(1); err == sqlite3.SQLITE_DONE {
			break
		} else if err != nil {
			t.Fatal(err)
		} else {
			t.Log(backup)
		}
	}
}
