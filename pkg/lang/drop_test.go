package lang_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_Drop_000(t *testing.T) {
	tests := []struct {
		In    SQStatement
		Query string
	}{
		{N("foo").DropTable(), `DROP TABLE foo`},
		{N("foo").WithSchema("main").DropTable(), `DROP TABLE main.foo`},
		{N("foo").WithSchema("main").DropTable().IfExists(), `DROP TABLE IF EXISTS main.foo`},
		{N("foo").DropView().IfExists(), `DROP VIEW IF EXISTS foo`},
		{N("foo").DropIndex().IfExists(), `DROP INDEX IF EXISTS foo`},
		{N("foo").DropTrigger().IfExists(), `DROP TRIGGER IF EXISTS foo`},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("Unexpected return from Query(): %q, wanted %q", v, test.Query)
		}
	}
}
