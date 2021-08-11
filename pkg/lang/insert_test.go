package lang_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_Insert_000(t *testing.T) {
	// Check db.P and db.V (value)
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{N("foo").Insert(), `INSERT INTO foo DEFAULT VALUES`, ``},
		{N("foo").WithSchema("main").Insert(), `INSERT INTO main.foo DEFAULT VALUES`, ``},
		{N("foo").WithSchema("main").Insert("a"), `INSERT INTO main.foo (a) VALUES (?)`, ``},
		{N("foo").WithSchema("main").Insert("a", "b"), `INSERT INTO main.foo (a,b) VALUES (?,?)`, ``},
		{N("foo").Replace(), `REPLACE INTO foo DEFAULT VALUES`, ``},
		{N("foo").WithSchema("main").Replace(), `REPLACE INTO main.foo DEFAULT VALUES`, ``},
		{N("foo").WithSchema("main").Replace("a"), `REPLACE INTO main.foo (a) VALUES (?)`, ``},
		{N("foo").WithSchema("main").Replace("a", "b"), `REPLACE INTO main.foo (a,b) VALUES (?,?)`, ``},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("db.V = %v, wanted %v", v, test.String)
		}
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("db.V = %v, wanted %v", v, test.Query)
			}
		}
	}
}
