package lang_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_Query_000(t *testing.T) {
	tests := []struct {
		In    SQStatement
		Query string
	}{
		{Q(""), ``},
		{Q("PRAGMA TEST"), `PRAGMA TEST`},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("got %v, wanted %v", v, test.Query)
		}
	}
}
