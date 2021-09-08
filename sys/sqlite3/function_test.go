package sqlite3_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/djthorpe/go-sqlite/sys/sqlite3"
)

func Test_Func_001(t *testing.T) {
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

	db.SetBusyTimeout(time.Second)

	// Create a function which sleeps
	if err := db.CreateScalarFunction("sleepy", 0, true, func(ctx *sqlite3.Context, args []*sqlite3.Value) {
		sqlite3.Sleep(time.Second * 5)
	}); err != nil {
		t.Error(err)
	}

	// Execute sleepy function
	if st, err := db.Prepare(fmt.Sprint("SELECT SLEEPY()")); err != nil {
		t.Error(err)
	} else if r, err := st.Exec(); err != nil {
		t.Error(err)
	} else {
		t.Log(r)
		for {
			row, err := r.Next()
			if err != nil {
				t.Error(err)
				break
			} else if row == nil {
				break
			}
			t.Log(row)
		}
	}
}
