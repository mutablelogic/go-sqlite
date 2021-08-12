package lang_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_Query_000(t *testing.T) {
	// Check db.P and db.V (value)
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{Q(""), `SELECT NULL`, `SELECT NULL`},
		{Q("PRAGMA TEST"), `PRAGMA TEST`, `PRAGMA TEST`},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("db.Q = %v, wanted %v", v, test.String)
		}
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("db.Q = %v, wanted %v", v, test.Query)
			}
		}
	}
}
