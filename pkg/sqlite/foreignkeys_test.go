package sqlite_test

import (
	"testing"

	sq "github.com/djthorpe/go-sqlite/pkg/sqlite"
)

func Test_ForeignKeys_001(t *testing.T) {
	db, err := sq.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.SetForeignKeyConstraints(true); err != nil {
		t.Error(err)
	} else if v, err := db.ForeignKeyConstraints(); err != nil {
		t.Error(err)
	} else if v != true {
		t.Error("Unexpected response from ForeignKeyConstraints")
	}

	if err := db.SetForeignKeyConstraints(false); err != nil {
		t.Error(err)
	} else if v, err := db.ForeignKeyConstraints(); err != nil {
		t.Error(err)
	} else if v != false {
		t.Error("Unexpected response from ForeignKeyConstraints")
	}
}
