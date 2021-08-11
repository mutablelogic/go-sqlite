package lang_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_Column_000(t *testing.T) {
	// Check db.C (column)
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{C("a"), `a TEXT`, ``},
		{C("a").WithType("BLOB"), `a BLOB`, ``},
		{C("a").NotNull(), `a TEXT NOT NULL`, ``},
		{C("a").WithType("VARCHAR"), `a "VARCHAR"`, ``},
		{C("a").WithAlias("b"), `a AS b`, ``},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("db.N = %v, wanted %v", v, test.String)
		}
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("db.N = %v, wanted %v", v, test.Query)
			}
		}
	}
}
