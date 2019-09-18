package sqlite_test

import (
	"strings"
	"testing"
	"time"

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

func Test_Util_004(t *testing.T) {
	var tests = []struct {
		from, decltype string
		notnull        bool
		to             interface{}
	}{
		{"", "TEXT", false, nil},
		{"", "TEXT", true, ""},
		{" ", "TEXT", false, " "},
		{" test ", "TEXT", false, " test "},
		{" test ", "TEXT", true, " test "},
		{"", "BOOL", false, nil},
		{"0", "BOOL", false, false},
		{"1", "BOOL", false, true},
		{"t", "BOOL", false, true},
		{"f", "BOOL", false, false},
		{" true ", "BOOL", false, true},
		{" false ", "BOOL", false, false},
		{"0", "INTEGER", false, int64(0)},
		{" -1 ", "INTEGER", false, int64(-1)},
		{"  ", "INTEGER", false, nil},
		{" 12345 ", "INTEGER", false, int64(12345)},
		{"0", "FLOAT", false, float64(0)},
		{" -1 ", "FLOAT", false, float64(-1)},
		{"  ", "FLOAT", false, nil},
		{" 123.45 ", "FLOAT", false, float64(123.45)},
		{" 123E45 ", "FLOAT", false, float64(123E45)},
		{" 1/1/70 ", "DATETIME", false, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{" 24/12/70 ", "DATETIME", false, time.Date(1970, 12, 24, 0, 0, 0, 0, time.UTC)},
		{" 24/12/50 ", "DATETIME", false, time.Date(1950, 12, 24, 0, 0, 0, 0, time.UTC)},
		{" 24/12/00 ", "DATETIME", false, time.Date(1900, 12, 24, 0, 0, 0, 0, time.UTC)},
		{" 24/12/2001 ", "DATETIME", false, time.Date(2001, 12, 24, 0, 0, 0, 0, time.UTC)},
	}
	for i, test := range tests {
		if result, err := sqlite.ToSupportedType(test.from, test.decltype, test.notnull, time.UTC); err != nil {
			t.Errorf("Test %v: %v error=%v", i, test.from, err)
		} else if result != test.to {
			t.Errorf("Test %v: Got %v, expected %v", i, result, test.to)
		}
	}
}

func Test_Util_005(t *testing.T) {
	var tests = []struct {
		from string
		to   string
	}{
		{"", "TEXT"},
		{"0", "BOOL INTEGER FLOAT TEXT"},
		{"1", "BOOL INTEGER FLOAT TEXT"},
		{"t", "BOOL TEXT"},
		{"f", "BOOL TEXT"},
		{"TRUE", "BOOL TEXT"},
		{"FALSE", "BOOL TEXT"},
		{"-1.5", "FLOAT TEXT"},
		{"1E5", "FLOAT TEXT"},
		{"01/01/2019", "DATETIME TEXT"},
		{"1 Jan 2019", "DATETIME TEXT"},
		{"12/29/2019", "TEXT"},
		{"29/12/19", "DATETIME TEXT"},
		{"1234EF", "BLOB TEXT"},
		{"1234E1", "FLOAT BLOB TEXT"},
		{"  2006-01-02T15:04:05Z ", "TIMESTAMP DATETIME TEXT"},
		{"  2006-01-02T15:04:05.9999Z ", "TIMESTAMP DATETIME TEXT"},
	}
	for i, test := range tests {
		decltypes := sqlite.SupportedTypesForValue(test.from)
		if strings.Join(decltypes, " ") != test.to {
			t.Errorf("Test %v: Got %v, expected %v for %v", i, strings.Join(decltypes, " "), test.to, test.from)
		}
	}
}
