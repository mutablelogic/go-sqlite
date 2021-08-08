package sqlite_test

import (
	"testing"

	sq "github.com/djthorpe/go-sqlite/pkg/sqlite"
)

func Test_001(t *testing.T) {
	if version := sq.Version(); version == "" {
		t.Fatal("Unexpected version")
	} else {
		t.Log("version=", version)
	}
}
