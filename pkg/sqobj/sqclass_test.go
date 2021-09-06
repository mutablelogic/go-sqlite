package sqobj_test

import (
	"testing"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"

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
	conn, err := sqlite.New()
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	class, err := RegisterClass(N("test"), TestClassStructA{A: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := class.Create(conn, "", SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Error(err)
	} else if err := class.Create(conn, "", SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Error(err)
	} else {
		t.Log(conn.Tables())
	}
}

///////////////////////////////////////////////////////////////////////////////
// TYPES

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

	db, err := sqlite.New()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	if err := cFile.Create(db, "main", SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Error(err)
	} else if err := cFileMark.Create(db, "main", SQLITE_FLAG_DELETEIFEXISTS); err != nil {
		t.Error(err)
	} else {
		t.Log(db.Tables())
	}
}
