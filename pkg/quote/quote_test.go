package quote_test

import (
	"testing"

	// Import Namespace
	. "github.com/djthorpe/go-sqlite/pkg/quote"
)

func Test_Quote_001(t *testing.T) {
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

func Test_Quote_002(t *testing.T) {
	var tests = []struct{ from, to string }{
		{"", `''`},
		{"test", `'test'`},
		{"test'", `'test'''`},
		{"'test'", `'''test'''`},
	}
	for i, test := range tests {
		if Quote(test.from) != test.to {
			t.Errorf("%d: Expected %s, got %s", i, test.to, Quote(test.from))
		}
	}
}

func Test_Quote_003(t *testing.T) {
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

func Test_Quote_004(t *testing.T) {
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
