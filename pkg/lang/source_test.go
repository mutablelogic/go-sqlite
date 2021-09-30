package lang_test

import (
	"fmt"
	"testing"

	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

type Statement interface {
	Query() string
}

func Test_Source_000(t *testing.T) {
	// Check db.N (name)
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{N("a"), `a`, `SELECT * FROM a`},
		{N("a").WithAlias("b"), `a AS b`, `SELECT * FROM a AS b`},
		{N("a").WithSchema("main"), `main.a`, `SELECT * FROM main.a`},
		{N("a").WithSchema("main").WithAlias("b"), `main.a AS b`, `SELECT * FROM main.a AS b`},
		{N("x y").WithSchema("main").WithAlias("b"), `main."x y" AS b`, `SELECT * FROM main."x y" AS b`},
		{N("insert").WithSchema("main").WithAlias("b"), `main."insert" AS b`, `SELECT * FROM main."insert" AS b`},
		{N("x").WithType("TEXT"), `x TEXT`, ``},
		{N("x").WithDesc(), `x DESC`, ``},
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
