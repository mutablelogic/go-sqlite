package lang_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_Delete_000(t *testing.T) {
	tests := []struct {
		In    SQStatement
		Query string
	}{
		{N("foo").Delete(nil), `DELETE FROM foo WHERE NULL`},
	}

	for i, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("Test %d, Unexpected return from Query(): %q, wanted %q", i, v, test.Query)
		}
	}
}
