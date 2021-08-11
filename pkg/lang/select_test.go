package lang_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_Select_000(t *testing.T) {
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{S(), `SELECT NULL`, ``},
		{S(N("a")), `SELECT * FROM a`, ``},
		{S(N("a"), N("b")), `SELECT * FROM a,b`, ``},
		{S(N("a").WithAlias("aa"), N("b").WithAlias("bb")), `SELECT * FROM a AS aa,b AS bb`, ``},
		{S(N("a").WithSchema("main").WithAlias("aa"), N("b").WithAlias("bb").WithSchema("main")), `SELECT * FROM main.a AS aa,main.b AS bb`, ``},
		{S(N("a")).WithDistinct(), `SELECT DISTINCT * FROM a`, ``},
		{S(N("a").WithAlias("aa"), N("b").WithAlias("bb")), "SELECT * FROM a AS aa,b AS bb", ""},
		{S(N("a")).WithLimitOffset(1, 0), "SELECT * FROM a LIMIT 1", ""},
		{S(N("a")).WithLimitOffset(0, 1), "SELECT * FROM a OFFSET 1", ""},
		{S(N("a")).WithLimitOffset(1, 1), "SELECT * FROM a LIMIT 1,1", ""},
		{S(N("a")).Where(nil), "SELECT * FROM a WHERE NULL", ""},
		{S(N("a")).Where(nil, nil), "SELECT * FROM a WHERE NULL AND NULL", ""},
		{S(N("a")).Where(N("a")), "SELECT * FROM a WHERE a", ""},
		{S(N("a")).Where(N("b"), N("c")), "SELECT * FROM a WHERE b AND c", ""},
		{S(N("a")).Where(P, P), "SELECT * FROM a WHERE ? AND ?", ""},
		{S(N("a")).Where(P).Where(P), "SELECT * FROM a WHERE ? AND ?", ""},
		{S(N("a")).Where(V("foo"), V(true)), "SELECT * FROM a WHERE 'foo' AND TRUE", ""},
		{S(N("a")).Where(V("foo"), V(false)), "SELECT * FROM a WHERE 'foo' AND FALSE", ""},
	}

	for i, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("Test %d, Unexpected return from String(): %q, wanted %q", i, v, test.String)
		} else {
			t.Log(v)
		}
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("Test %d, Unexpected return from Query(): %q, wanted %q", i, v, test.Query)
			}
		}
	}
}
