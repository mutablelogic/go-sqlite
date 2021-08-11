package lang_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_Expr_000(t *testing.T) {
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{P, `?`, `SELECT ?`},
		{V(nil), `NULL`, `SELECT NULL`},
		{V("test"), `'test'`, `SELECT 'test'`},
		{V("te\"st"), `'te"st'`, `SELECT 'te"st'`},
		{V("te'st"), `'te''st'`, `SELECT 'te''st'`},
		{V(true), `TRUE`, `SELECT TRUE`},
		{V(false), `FALSE`, `SELECT FALSE`},
		{V(0), `0`, `SELECT 0`},
		{V(-1), `-1`, `SELECT -1`},
		{V(1.1), `1.1`, `SELECT 1.1`},
		{V(-1.1), `-1.1`, `SELECT -1.1`},
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
func Test_Expr_001(t *testing.T) {
	// Check db.P and db.V (value)
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{V(nil).Or(nil), `NULL OR NULL`, ``},
		{V(100).Or(200), `100 OR 200`, ``},
		{V(`a`).Or(`b`), `'a' OR 'b'`, ``},
		{V(100).Or(200).Or(300), `100 OR 200 OR 300`, ``},
		{V(N(`a`)).Or(`b`), `a OR 'b'`, ``},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("db.V = %q, wanted %q", v, test.String)
		}
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("db.V = %q, wanted %q", v, test.Query)
			}
		}
	}
}
