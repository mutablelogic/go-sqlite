package lang_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_Update_000(t *testing.T) {
	tests := []struct {
		In    SQStatement
		Query string
	}{
		{N("foo").Update(), `UPDATE foo`},
		{N("foo").Update().WithFail(), `UPDATE OR FAIL foo`},
		{N("foo").Update().WithAbort(), `UPDATE OR ABORT foo`},
		{N("foo").Update().WithIgnore(), `UPDATE OR IGNORE foo`},
		{N("foo").Update().WithReplace(), `UPDATE OR REPLACE foo`},
		{N("foo").Update().WithRollback(), `UPDATE OR ROLLBACK foo`},
		{N("foo").Update("bar"), `UPDATE foo SET bar=?`},
		{N("foo").Update("bar", "baz"), `UPDATE foo SET bar=?, baz=?`},
		{N("foo").Update("bar").Where(Q("baz IS NULL")), `UPDATE foo SET bar=? WHERE baz IS NULL`},
		{N("foo").Update("bar").Where(Q("baz IS NULL"), Q("baz IS NOT NULL")), `UPDATE foo SET bar=? WHERE baz IS NULL AND baz IS NOT NULL`},
	}

	for i, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("Test %d, Unexpected return from Query(): %q, wanted %q", i, v, test.Query)
		}
	}
}
