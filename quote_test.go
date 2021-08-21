package sqlite_test

import (
	"testing"

	"github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite"
)

func Test_Util_001(t *testing.T) {
	var tests = []struct{ from, to string }{
		{"", `""`},
		{"test", `"test"`},
		{"test\"", `"test"""`},
		{"\"test\"", `"""test"""`},
	}
	for i, test := range tests {
		if DoubleQuote(test.from) != test.to {
			t.Errorf("%d: Expected %s, got %s", i, test.to, DoubleQuote(test.from))
		}
	}
}

func Test_Util_002(t *testing.T) {
	var tests = []struct{ from, to string }{
		{"", `""`},
		{"test", `test`},
		{"temp", `"temp"`},
		{"some other", `"some other"`},
	}
	for i, test := range tests {
		if QuoteIdentifier(test.from) != test.to {
			t.Errorf("%d: Expected %s, got %s", i, test.to, QuoteIdentifier(test.from))
		}
	}
}

func Test_Util_003(t *testing.T) {
	var tests = []struct {
		from []string
		to   string
	}{
		{[]string{""}, `""`},
		{[]string{"test"}, `test`},
		{[]string{"temp"}, `"temp"`},
		{[]string{"some other"}, `"some other"`},
		{[]string{"test", "temp"}, `test,"temp"`},
		{[]string{"temp", "test"}, `"temp",test`},
		{[]string{"some other", "select"}, `"some other","select"`},
	}
	for i, test := range tests {
		if QuoteIdentifiers(test.from...) != test.to {
			t.Errorf("%d: Expected %s, got %s", i, test.to, QuoteIdentifiers(test.from...))
		}
	}
}

func Test_Util_004(t *testing.T) {
	ts := sqlite.SupportedTypes()
	for _, ts := range ts {
		if sqlite.IsSupportedType(ts) == false {
			t.Error(ts, ": Expected supported type, got unsupported type")
		}
	}
}
func Test_Util_005(t *testing.T) {
	var tests = []struct{ from, to string }{
		{"", `''`},
		{"test", `'test'`},
		{"test\"", `'test"'`},
		{"'test'", `'''test'''`},
	}
	for i, test := range tests {
		if Quote(test.from) != test.to {
			t.Errorf("%d: Expected %s, got %s", i, test.to, Quote(test.from))
		}
	}
}
