package sqlite3_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
	"unsafe"

	"github.com/djthorpe/go-sqlite/sys/sqlite3"
)

func Test_Trace_001(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	db, err := sqlite3.OpenPathEx(filepath.Join(tmpdir, "test.sqlite"), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	}

	db.SetTraceHook(func(t sqlite3.TraceType, a, b unsafe.Pointer) int {
		switch t {
		case sqlite3.SQLITE_TRACE_STMT:
			fmt.Println("TRACE => ", t, (*sqlite3.Statement)(a), b)
		case sqlite3.SQLITE_TRACE_ROW:
			fmt.Println("TRACE => ", t, (*sqlite3.Statement)(a))
		case sqlite3.SQLITE_TRACE_PROFILE:
			ms := time.Duration(time.Duration(*(*int64)(b)) * time.Millisecond)
			fmt.Println("TRACE => ", t, (*sqlite3.Statement)(a), ms)
		case sqlite3.SQLITE_TRACE_CLOSE:
			fmt.Println("TRACE => ", t, (*sqlite3.Conn)(a))
		}
		return 0
	}, 0xFF)

	if err := db.Exec("CREATE TABLE test (a TEXT)", nil); err != nil {
		t.Error(err)
	}

	if err := db.Close(); err != nil {
		t.Error(err)
	}

	db.SetTraceHook(nil, 0)
}
