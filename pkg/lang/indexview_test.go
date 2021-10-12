package lang_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_CreateIndexView_000(t *testing.T) {
	tests := []struct {
		In    SQStatement
		Query string
	}{
		{N("foo").CreateIndex("foo"), `CREATE INDEX foo ON foo ()`},
		{N("foo").CreateIndex("foo").IfNotExists(), `CREATE INDEX IF NOT EXISTS foo ON foo ()`},
		{N("foo").CreateIndex("bar").WithUnique(), `CREATE UNIQUE INDEX foo ON bar ()`},
		{N("foo").CreateIndex("bar", "a", "b").WithUnique(), `CREATE UNIQUE INDEX foo ON bar (a,b)`},
		{N("foo").WithSchema("main").CreateIndex("bar", "a", "b").WithUnique(), `CREATE UNIQUE INDEX main.foo ON bar (a,b)`},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("Unexpected return from Query(): %q, wanted %q", v, test.Query)
		}
	}
}

func Test_CreateIndexView_001(t *testing.T) {
	tests := []struct {
		In    SQStatement
		Query string
	}{
		{N("foo").CreateVirtualTable("bar"), `CREATE VIRTUAL TABLE foo USING bar`},
		{N("foo").WithSchema("main").CreateVirtualTable("bar"), `CREATE VIRTUAL TABLE main.foo USING bar`},
		{N("foo").CreateVirtualTable("bar").IfNotExists(), `CREATE VIRTUAL TABLE IF NOT EXISTS foo USING bar`},
		{N("foo").CreateVirtualTable("bar", "a", "b"), `CREATE VIRTUAL TABLE foo USING bar (a,b)`},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("Unexpected return from Query(): %q, wanted %q", v, test.Query)
		}
	}
}

func Test_CreateIndexView_002(t *testing.T) {
	tests := []struct {
		In    SQStatement
		Query string
	}{
		{N("foo").CreateView(nil, "bar"), `CREATE VIEW foo (bar)`},
		{N("foo").CreateView(nil, "bar").IfNotExists(), `CREATE VIEW IF NOT EXISTS foo (bar)`},
		{N("foo").CreateView(nil, "bar").WithTemporary(), `CREATE TEMPORARY VIEW foo (bar)`},
		{N("foo").CreateView(nil, "col1", "col2").WithTemporary(), `CREATE TEMPORARY VIEW foo (col1,col2)`},
		{N("foo").CreateView(S(N("x"))), `CREATE VIEW foo AS SELECT * FROM x`},
		{N("foo").CreateView(S(N("x")), "rowid"), `CREATE VIEW foo (rowid) AS SELECT * FROM x`},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("Unexpected return from Query(): %q, wanted %q", v, test.Query)
		}
	}
}
