package lang_test

import (
	"fmt"
	"testing"

	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

func Test_Delete_000(t *testing.T) {
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{N("foo").Delete(nil), `DELETE FROM foo WHERE NULL`, ``},
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
