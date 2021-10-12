package lang_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_Join_000(t *testing.T) {
	tests := []struct {
		In     SQExpr
		String string
	}{
		{J(N("foo"), N("bar")), `foo CROSS JOIN bar`},
		{J(N("foo"), N("bar")).Join(), `foo JOIN bar`},
		{J(N("foo"), N("bar")).LeftJoin(), `foo LEFT JOIN bar`},
		{J(N("foo"), N("bar")).LeftInnerJoin(), `foo LEFT INNER JOIN bar`},
		{J(N("foo").WithAlias("a"), N("bar").WithAlias("b")).Join(), `foo AS a JOIN bar AS b`},
		{J(N("foo"), N("bar")).Join(Q("a=b")), `foo JOIN bar ON a=b`},
		{J(N("foo"), N("bar")).Join(Q("a=b"), Q("b=c")), `foo JOIN bar ON a=b AND b=c`},
	}

	for _, test := range tests {
		if v := test.In.String(); v != test.String {
			t.Errorf("got %q, wanted %q", v, test.String)
		} else {
			t.Logf(v)
		}
	}
}

func Test_Join_001(t *testing.T) {
	tests := []struct {
		In     SQExpr
		String string
	}{
		{J(N("foo"), N("bar")).Join().Using("a", "b"), `foo JOIN bar USING (a,b)`},
	}

	for _, test := range tests {
		if v := test.In.String(); v != test.String {
			t.Errorf("got %q, wanted %q", v, test.String)
		} else {
			t.Logf(v)
		}
	}
}
