package sqlite_test

import (
	"testing"

	sq "github.com/djthorpe/go-sqlite/pkg/sqlite"
	"github.com/djthorpe/sqlite"
)

func Test_Util_001(t *testing.T) {
	var tests = []struct{ from, to string }{
		{"", `""`},
		{"test", `"test"`},
		{"test\"", `"test"""`},
		{"\"test\"", `"""test"""`},
	}
	for i, test := range tests {
		if sq.DoubleQuote(test.from) != test.to {
			t.Errorf("%d: Expected %s, got %s", i, test.to, sq.DoubleQuote(test.from))
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
		if sq.QuoteIdentifier(test.from) != test.to {
			t.Errorf("%d: Expected %s, got %s", i, test.to, sq.QuoteIdentifier(test.from))
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
		if sq.QuoteIdentifiers(test.from...) != test.to {
			t.Errorf("%d: Expected %s, got %s", i, test.to, sq.QuoteIdentifiers(test.from...))
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
