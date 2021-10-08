package lang_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_Trigger_000(t *testing.T) {
	tests := []struct {
		In    SQStatement
		Query string
	}{
		{N("a").CreateTrigger("b", Q("statement_a")), `CREATE TRIGGER a AFTER INSERT ON b BEGIN statement_a; END`},
		{N("a").WithSchema("s").CreateTrigger("b", Q("statement_a")), `CREATE TRIGGER s.a AFTER INSERT ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).IfNotExists(), `CREATE TRIGGER IF NOT EXISTS a AFTER INSERT ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).WithTemporary(), `CREATE TEMPORARY TRIGGER a AFTER INSERT ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).Before().Insert(), `CREATE TRIGGER a BEFORE INSERT ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).After().Insert(), `CREATE TRIGGER a AFTER INSERT ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).InsteadOf().Insert(), `CREATE TRIGGER a INSTEAD OF INSERT ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).Before().Delete(), `CREATE TRIGGER a BEFORE DELETE ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).After().Delete(), `CREATE TRIGGER a AFTER DELETE ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).InsteadOf().Delete(), `CREATE TRIGGER a INSTEAD OF DELETE ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).Before().Update(), `CREATE TRIGGER a BEFORE UPDATE ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).After().Update(), `CREATE TRIGGER a AFTER UPDATE ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).InsteadOf().Update(), `CREATE TRIGGER a INSTEAD OF UPDATE ON b BEGIN statement_a; END`},
		{N("a").CreateTrigger("b", Q("statement_a")).Before().Update("x", "y"), `CREATE TRIGGER a BEFORE UPDATE OF (x,y) ON b BEGIN statement_a; END`},
		{N("a").WithSchema("s").CreateTrigger("b", Q("statement_a"), Q("statement_b")), `CREATE TRIGGER s.a AFTER INSERT ON b BEGIN statement_a; statement_b; END`},
	}

	for _, test := range tests {
		if test.In == nil {
			t.Errorf("Trigger: returned nil for %q", test.Query)
		} else if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("Trigger: returned %q, wanted %q", v, test.Query)
			}
		}
	}
}
