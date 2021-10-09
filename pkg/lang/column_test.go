package lang_test

import (
	"testing"

	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_Column_000(t *testing.T) {
	// Check db.C (column)
	tests := []struct {
		In     SQExpr
		String string
	}{
		{C("a"), `a TEXT`},
		{C("a").WithType("BLOB"), `a BLOB`},
		{C("a").NotNull(), `a TEXT NOT NULL`},
		{C("a").WithType("VARCHAR"), `a VARCHAR`},
		{C("a").WithAlias("b"), `a AS b`},
	}

	for _, test := range tests {
		if v := test.In.String(); v != test.String {
			t.Errorf("db.N = %v, wanted %v", v, test.String)
		}
	}
}
