package sqobj_test

import (
	"testing"
	"time"

	// Modules
	sqlite3 "github.com/djthorpe/go-sqlite/pkg/sqlite3"

	// Namespace importst
	. "github.com/djthorpe/go-sqlite/pkg/lang"
	. "github.com/djthorpe/go-sqlite/pkg/sqobj"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type File struct {
	Index   string    `sqlite:"index,primary" json:"index"`
	Path    string    `sqlite:"path,primary" json:"path"`
	Name    string    `sqlite:"name,primary" json:"name"`
	IsDir   bool      `sqlite:"is_dir,not null" json:"is_dir"`
	Ext     string    `sqlite:"ext" json:"ext"`
	ModTime time.Time `sqlite:"modtime" json:"modtime"`
	Size    int64     `sqlite:"size,not null" json:"size"`
}

type FileMark struct {
	Index     string    `sqlite:"index,primary,foreign"`
	Path      string    `sqlite:"path,primary,foreign"`
	Name      string    `sqlite:"name,primary,foreign"`
	Mark      bool      `sqlite:"mark,not null"`
	IndexTime time.Time `sqlite:"idxtime"`
}

func Test_Objects_001(t *testing.T) {
	conn, err := sqlite3.New()
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	// Set up tracing function
	conn.SetTraceHook(func(sql string, d time.Duration) {
		if d >= 0 {
			t.Log("EXEC:", sql, "=>", d)
		}
	})

	// Register classes
	cFile := MustRegisterClass(N("file"), File{})
	cFileMark := MustRegisterClass(N("filemark"), FileMark{}).ForeignKey(cFile)

	// Make database and ensure cFile and cFileMark are registered
	obj, err := With(conn, "main", cFile, cFileMark)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(obj)
	}
}
