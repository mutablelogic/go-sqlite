package sqlite3_test

import (
	"io"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/mutablelogic/go-sqlite/sys/sqlite3"
)

func Test_Results_001(t *testing.T) {
	db, err := sqlite3.OpenPathEx(":memory:", sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	st, err := db.Prepare("SELECT ?")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Now()

	var tests = []struct {
		in, out interface{}
	}{
		{int(1), int64(1)},
		{int8(2), int64(2)},
		{int16(3), int64(3)},
		{int32(4), int64(4)},
		{int64(5), int64(5)},
		{uint(6), int64(6)},
		{uint8(7), int64(7)},
		{uint16(8), int64(8)},
		{uint32(9), int64(9)},
		{uint64(10), int64(10)},
		{"test", "test"},
		{false, int64(0)},
		{true, int64(1)},
		{float64(math.Pi), float64(math.Pi)},
		{float32(math.Pi), float64(float32(math.Pi))},
		{now, now.Format(time.RFC3339)},
		{time.Time{}, nil},
		{nil, nil},
	}

	for _, test := range tests {
		r, err := st.Exec(0, test.in)
		if err != nil {
			t.Fatal(err)
		}
		for {
			values, err := r.Next()
			if err == io.EOF {
				break
			}
			if len(values) != 1 {
				t.Error("Data count should be one")
			}
			out := values[0]
			if out != test.out {
				t.Errorf("Expected %v (%T) but got %v (%T) for bind type %T", test.out, test.out, out, out, test.in)
			}
		}
	}
}

func Test_Results_002(t *testing.T) {
	db, err := sqlite3.OpenPathEx(":memory:", sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	st, err := db.Prepare("SELECT ?")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	now := time.Now()

	var tests = []struct {
		in, out interface{}
	}{
		{int(1), int(1)},
		{int8(2), int8(2)},
		{int16(3), int16(3)},
		{int32(4), int32(4)},
		{int64(5), int64(5)},
		{uint(6), uint(6)},
		{uint8(7), uint8(7)},
		{uint16(8), uint16(8)},
		{uint32(9), uint32(9)},
		{uint64(10), uint64(10)},
		{"test", "test"},
		{false, false},
		{true, true},
		{float64(math.Pi), float64(math.Pi)},
		{float32(math.Pi), float32(math.Pi)},
		{now, now.Truncate(time.Second)},
		{time.Time{}, time.Time{}},
	}

	for _, test := range tests {
		r, err := st.Exec(0, test.in)
		if err != nil {
			t.Fatal(err)
		}
		for {
			values, err := r.Next(reflect.TypeOf(test.out))
			if err == io.EOF {
				break
			}
			if len(values) != 1 {
				t.Error("Data count should be one")
			}
			out := values[0]
			if out != test.out {
				t.Errorf("Expected %v (%T) but got %v (%T) for bind type %T", test.out, test.out, out, out, test.in)
			}
		}
	}
}
