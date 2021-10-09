package lang_test

import (
	"fmt"
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_Expr_000(t *testing.T) {
	tests := []struct {
		In     SQExpr
		String string
	}{
		{P, `?`},
		{V(nil), `NULL`},
		{V("test"), `'test'`},
		{V("te\"st"), `'te"st'`},
		{V("te'st"), `'te''st'`},
		{V(true), `TRUE`},
		{V(false), `FALSE`},
		{V(0), `0`},
		{V(-1), `-1`},
		{V(1.1), `1.1`},
		{V(-1.1), `-1.1`},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("Unexpected return from String(): %q, wanted %q", v, test.String)
		}
	}
}
