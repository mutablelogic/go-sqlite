package sqlite3_test

import (
	"math"
	"testing"
	"time"

	"github.com/mutablelogic/go-sqlite/sys/sqlite3"
)

func Test_Bind_001(t *testing.T) {
	db, err := sqlite3.OpenPath(":memory:", sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	st, _, err := db.Prepare("SELECT ?")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Finalize()

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
		st.Reset()
		st.Bind(test.in)
		for st.Step() == sqlite3.SQLITE_ROW {
			if st.DataCount() != 1 {
				t.Error("Data count should be one")
			}
			out := st.ColumnInterface(0)
			if out != test.out {
				t.Errorf("Expected %v (%T) but got %v (%T) for bind type %T", test.out, test.out, out, out, test.in)
			}
		}
	}
}
