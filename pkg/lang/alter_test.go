package lang_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_Alter_000(t *testing.T) {
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{N("foo").AlterTable(), `ALTER TABLE foo`, ``},
		{N("foo").WithSchema("main").AlterTable(), `ALTER TABLE main.foo`, ``},
		{N("foo").WithSchema("main").AlterTable().DropColumn(C("a")), `ALTER TABLE main.foo DROP COLUMN a`, ``},
		{N("foo").WithSchema("main").AlterTable().AddColumn(C("a")), `ALTER TABLE main.foo ADD COLUMN a TEXT`, ``},
		{N("foo").WithSchema("main").AlterTable().AddColumn(C("a").NotNull()), `ALTER TABLE main.foo ADD COLUMN a TEXT NOT NULL`, ``},
		{N("foo").WithSchema("main").AlterTable().AddColumn(C("a").WithPrimary()), `ALTER TABLE main.foo ADD COLUMN a TEXT NOT NULL PRIMARY KEY`, ``},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("Unexpected return from String(): %q, wanted %q", v, test.String)
		}
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("Unexpected return from Query(): %q, wanted %q", v, test.Query)
			}
		}
	}
}
