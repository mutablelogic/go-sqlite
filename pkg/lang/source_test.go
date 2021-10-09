package lang_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_Source_000(t *testing.T) {
	tests := []struct {
		In     SQExpr
		String string
	}{
		{N("a"), `a`},
		{N("a").WithAlias("b"), `a AS b`},
		{N("a").WithSchema("main"), `main.a`},
		{N("a").WithSchema("main").WithAlias("b"), `main.a AS b`},
		{N("x y").WithSchema("main").WithAlias("b"), `main."x y" AS b`},
		{N("insert").WithSchema("main").WithAlias("b"), `main."insert" AS b`},
		{N("x").WithType("TEXT"), `x TEXT`},
		{N("x").WithDesc(), `x DESC`},
	}

	for _, test := range tests {
		if v := test.In.String(); v != test.String {
			t.Errorf("got %v, wanted %v", v, test.String)
		}
	}
}
