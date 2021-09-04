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
func Test_Value_002(t *testing.T) {
	// Convert nil to zero value
	if out, err := sqlite.UnboundValue(nil, reflect.TypeOf(int(0))); err != nil {
		t.Error("Unexpected error:", err)
	} else if out.Int() != int64(0) {
		t.Errorf("Expected %v, got %v", int(0), out)
	}
	if out, err := sqlite.UnboundValue(nil, reflect.TypeOf(false)); err != nil {
		t.Error("Unexpected error:", err)
	} else if out.Bool() != false {
		t.Errorf("Expected %v, got %v", false, out)
	}
	if out, err := sqlite.UnboundValue(nil, reflect.TypeOf("")); err != nil {
		t.Error("Unexpected error:", err)
	} else if out.String() != "" {
		t.Errorf("Expected %q, got %q", "", out)
	}
	if out, err := sqlite.UnboundValue(nil, reflect.TypeOf(time.Time{})); err != nil {
		t.Error("Unexpected error:", err)
	} else if out.Interface().(time.Time).IsZero() == false {
		t.Errorf("Expected %v, got %v", time.Time{}, out)
	}
	// Convert to int
	if out, err := sqlite.UnboundValue(int64(100), reflect.TypeOf(int(0))); err != nil {
		t.Error("Unexpected error:", err)
	} else if out.Kind() != reflect.Int {
		t.Error("Unexpected kind:", out.Kind())
	} else {
		t.Log(out.Kind(), out)
	}
	// Convert to uint
	if out, err := sqlite.UnboundValue(int64(100), reflect.TypeOf(uint(0))); err != nil {
		t.Error("Unexpected error:", err)
	} else if out.Kind() != reflect.Uint {
		t.Error("Unexpected kind:", out.Kind())
	} else {
		t.Log(out.Kind(), out)
	}
	// Convert to bool
	if out, err := sqlite.UnboundValue(int64(1), reflect.TypeOf(false)); err != nil {
		t.Error("Unexpected error:", err)
	} else if out.Kind() != reflect.Bool {
		t.Error("Unexpected kind:", out.Kind())
	} else if out.Bool() != true {
		t.Error("Unexpected value:", out)
	}
	if out, err := sqlite.UnboundValue(int64(0), reflect.TypeOf(false)); err != nil {
		t.Error("Unexpected error:", err)
	} else if out.Kind() != reflect.Bool {
		t.Error("Unexpected kind:", out.Kind())
	} else if out.Bool() != false {
		t.Error("Unexpected value:", out)
	}
	// Convert to string
	if out, err := sqlite.UnboundValue("test", reflect.TypeOf("")); err != nil {
		t.Error("Unexpected error:", err)
	} else if out.Kind() != reflect.String {
		t.Error("Unexpected kind:", out.Kind())
	} else if out.String() != "test" {
		t.Error("Unexpected value:", out)
	}
}
