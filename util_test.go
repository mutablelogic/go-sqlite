package sqlite_test

import (
	"testing"

	// Frameworks
	"github.com/djthorpe/sqlite"
)

func Test_Util_001(t *testing.T) {
	t.Log("Test_Util_001")
}

func Test_Util_002(t *testing.T) {
	var tests = []struct{ from, to string }{
		{"", "\"\""},
		{"test", "\"test\""},
		{"test\"", "\"test\"\"\""},
	}
	for i, test := range tests {
		quoted := sqlite.DoubleQuote(test.from)
		if quoted != test.to {
			t.Errorf("Test %v: Got %v, expected %v", i, quoted, test.to)
		}
	}
}

func Test_Util_003(t *testing.T) {
	var tests = []struct{ from, to string }{
		{"", "\"\""},
		{"test", "test"},
		{"select", "\"select\""},
		{"select that", "\"select that\""},
		{"test that", "\"test that\""},
	}
	for i, test := range tests {
		quoted := sqlite.QuoteIdentifier(test.from)
		if quoted != test.to {
			t.Errorf("Test %v: Got %v, expected %v", i, quoted, test.to)
		}
	}
}
