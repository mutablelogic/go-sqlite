package lang_test

import (
	"testing"

	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_ForeignKey_000(t *testing.T) {
	tests := []struct {
		In     SQForeignKey
		String string
	}{
		{N("foo").ForeignKey(), `FOREIGN KEY (foo) REFERENCES foo`},
		{N("index").ForeignKey(), `FOREIGN KEY (foo) REFERENCES "index"`},
		{N("index").ForeignKey().OnDeleteCascade(), `FOREIGN KEY (foo) REFERENCES "index" ON DELETE CASCADE`},
		{N("index").ForeignKey("a", "b").OnDeleteCascade(), `FOREIGN KEY (foo) REFERENCES "index" (a,b) ON DELETE CASCADE`},
	}

	for i, test := range tests {
		if v := test.In.Query("foo"); v != test.String {
			t.Errorf("Test %d, Unexpected return from String(): %q, wanted %q", i, v, test.String)
		} else {
			t.Log(v)
		}
	}
}
