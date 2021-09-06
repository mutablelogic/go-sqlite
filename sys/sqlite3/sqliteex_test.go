package sqlite3_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/djthorpe/go-sqlite/sys/sqlite3"
)

func Test_SQLiteEx_001(t *testing.T) {
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

	for i := 0; i < 10; i++ {
		if st, err := db.Prepare(fmt.Sprint("SELECT ", i)); err != nil {
			t.Error(err)
		} else if r, err := st.Exec(); err != nil {
			t.Error(err)
		} else {
			t.Log(r)
			for {
				row := r.Next()
				if row == nil {
					break
				}
				t.Log(row)
			}
		}
	}
}
