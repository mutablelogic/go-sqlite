// Package test creates managers for tests, of temporary databases, which are
// closed, and deleted, when the tests end.
package test

import (
	"path/filepath"
	"testing"

	// Packages
	filesearch "github.com/mutablelogic/go-sqlite/filesearch"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Database is the name of the temporary database
	Database = "filesearch.db"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// New returns a manager, of a database, in a temporary folder, with the
// options, which is closed, and deleted, with its folder, when the test
// ends. It fails the test when it can't be created.
func New(t testing.TB, opts ...filesearch.Opt) *filesearch.Manager {
	t.Helper()
	manager, err := filesearch.New(t.Context(), Path(t), opts...)
	if err != nil {
		t.Fatalf("filesearch.New: %v", err)
	}
	t.Cleanup(func() {
		if err := manager.Close(); err != nil {
			t.Errorf("filesearch.Close: %v", err)
		}
	})
	return manager
}

// Path returns the path of a database, in a temporary folder, which is
// deleted, with the folder, when the test ends
func Path(t testing.TB) string {
	t.Helper()
	return filepath.Join(t.TempDir(), Database)
}
