package sqlite_test

import (
	"os"
	"path/filepath"
	"testing"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
)

func Test_Attach_001(t *testing.T) {
	tmp, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Attach("test", filepath.Join(tmp, "db.sqlite")); err != nil {
		t.Fatal(err)
	} else if schemas := db.Schemas(); len(schemas) != 2 {
		t.Error("Unexpected number of schemas", schemas)
	} else {
		for _, schema := range schemas {
			t.Logf("%s => %q", schema, db.Filename(schema))
		}
	}
	if err := db.Detach("test"); err != nil {
		t.Fatal(err)
	}
}
