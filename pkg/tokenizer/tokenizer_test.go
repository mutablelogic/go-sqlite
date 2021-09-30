package tokenizer_test

import (
	"testing"

	// Namespace Imports
	. "github.com/mutablelogic/go-sqlite/pkg/tokenizer"
)

func Test_Tokenizer_001(t *testing.T) {
	var tests = []string{
		"CREATE TABLE foo (id INTEGER PRIMARY KEY, name TEXT);",
		`INSERT INTO foo VALUES ('string value',99,"name value",CURRENT_TIMESTAMP)`,
	}
	for _, test := range tests {
		tokenizer := NewTokenizer(test)
		tokens := []interface{}{}
		for {
			token, err := tokenizer.Next()
			if token == nil {
				break
			}
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			tokens = append(tokens, token)
		}
		t.Logf("%q =>", test)
		for _, token := range tokens {
			t.Logf("  <%T %q>", token, token)
		}
	}
}
