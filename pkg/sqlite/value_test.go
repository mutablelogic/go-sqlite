package sqlite_test

import (
	"math"
	"reflect"
	"testing"
	"time"

	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
)

func Test_Value_001(t *testing.T) {
	now := time.Now()
	tests := []struct {
		in        interface{}
		out       interface{}
		expecterr bool
	}{
		{int(100), int64(100), false},
		{int8(100), int64(100), false},
		{int16(100), int64(100), false},
		{int32(100), int64(100), false},
		{int64(100), int64(100), false},
		{int64(math.MaxInt64), int64(math.MaxInt64), false},
		{uint(200), int64(200), false},
		{uint8(200), int64(200), false},
		{uint16(200), int64(200), false},
		{uint32(200), int64(200), false},
		{uint64(200), int64(200), false},
		{uint64(math.MaxInt64 + 1), nil, true},
		{nil, nil, false},
		{false, false, false},
		{true, true, false},
		{float32(math.E), float64(float32(math.E)), false},
		{float64(math.Pi), float64(math.Pi), false},
		{"hello, world", "hello, world", false},
		{now, now, false},
		{time.Time{}, nil, false},
	}
	for _, test := range tests {
		out, err := sqlite.BoundValue(reflect.ValueOf(test.in))
		if err != nil {
			if !test.expecterr {
				t.Error("Unexpected error:", err, "for", test.in)
			} else {
				t.Log("Got expected error:", err, "for", test.in)
			}
			continue
		}
		if out != test.out {
			t.Errorf("Expected %v, got %v", test.out, out)
		}
	}
}
