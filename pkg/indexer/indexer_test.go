package indexer_test

import (
	"os"
	"testing"

	"github.com/djthorpe/go-sqlite/pkg/indexer"
	"github.com/djthorpe/go-sqlite/pkg/sqlite"
)

func Test_Indexer_001(t *testing.T) {
	tmppath, err := os.MkdirTemp("", "indexer_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmppath)
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if indexer, err := indexer.NewManager(db, "main", 0); err != nil {
		t.Fatal(err)
	} else {
		t.Log(indexer)
	}

}
