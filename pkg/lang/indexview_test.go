package lang_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_CreateIndex_000(t *testing.T) {
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{N("foo").CreateIndex("foo"), `CREATE INDEX foo ON foo ()`, ``},
		{N("foo").CreateIndex("foo").IfNotExists(), `CREATE INDEX IF NOT EXISTS foo ON foo ()`, ``},
		{N("foo").CreateIndex("bar").WithUnique(), `CREATE UNIQUE INDEX foo ON bar ()`, ``},
		{N("foo").CreateIndex("bar", "a", "b").WithUnique(), `CREATE UNIQUE INDEX foo ON bar (a,b)`, ``},
		{N("foo").WithSchema("main").CreateIndex("bar", "a", "b").WithUnique(), `CREATE UNIQUE INDEX main.foo ON bar (a,b)`, ``},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("Unexpected return from String(): %q, wanted %q", v, test.String)
		} else {
			t.Log(v)
		}
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("Unexpected return from Query(): %q, wanted %q", v, test.Query)
			}
		}
	}
}
