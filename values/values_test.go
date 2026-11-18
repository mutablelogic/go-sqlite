package values_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	// Packages
	values "github.com/mutablelogic/go-sqlite/values"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// value is a value which SQLite gives, which converts as SQLite does
type value struct {
	typ   values.Type
	value any
}

func (v value) Type() values.Type { return v.typ }
func (v value) Int64() int64 {
	switch x := v.value.(type) {
	case int64:
		return x
	case float64:
		return int64(x)
	}
	return 0
}
func (v value) Float() float64 {
	switch x := v.value.(type) {
	case int64:
		return float64(x)
	case float64:
		return x
	}
	return 0
}
func (v value) Text() string {
	switch x := v.value.(type) {
	case nil:
		return ""
	case []byte:
		return string(x)
	}
	return fmt.Sprint(v.value)
}
func (v value) Blob() []byte { return []byte(v.Text()) }

func integer(n int64) value { return value{values.Integer, n} }
func float(f float64) value { return value{values.Float, f} }
func text(s string) value   { return value{values.Text, s} }
func blob(b []byte) value   { return value{values.Blob, b} }
func null() value           { return value{values.Null, nil} }

// setter records the value which is set
type setter struct {
	typ   values.Type
	value any
}

func (s *setter) SetNull() error           { s.typ, s.value = values.Null, nil; return nil }
func (s *setter) SetInt64(n int64) error   { s.typ, s.value = values.Integer, n; return nil }
func (s *setter) SetFloat(f float64) error { s.typ, s.value = values.Float, f; return nil }
func (s *setter) SetText(t string) error   { s.typ, s.value = values.Text, t; return nil }
func (s *setter) SetBlob(b []byte) error   { s.typ, s.value = values.Blob, b; return nil }

type name string
type id int32

// valuer is a driver.Valuer, of its value, or an error
type valuer struct {
	value any
	err   error
}

func (v valuer) Value() (driver.Value, error) { return v.value, v.err }

func Test_Set(t *testing.T) {
	str := "a"
	when := time.Date(2026, 11, 18, 12, 30, 0, 123, time.UTC)
	for _, test := range []struct {
		value any
		typ   values.Type
		want  any
	}{
		{nil, values.Null, nil},
		{(*string)(nil), values.Null, nil},
		{true, values.Integer, int64(1)},
		{false, values.Integer, int64(0)},
		{int8(-7), values.Integer, int64(-7)},
		{id(9), values.Integer, int64(9)},
		{uint64(7), values.Integer, int64(7)},
		{float32(1.5), values.Float, 1.5},
		{"a", values.Text, "a"},
		{name("b"), values.Text, "b"},
		{&str, values.Text, "a"},
		{[]byte{1}, values.Blob, []byte{1}},
		{when, values.Text, "2026-11-18T12:30:00.000000123Z"},
		{valuer{value: int64(3)}, values.Integer, int64(3)},
		{sql.NullString{}, values.Null, nil},
	} {
		var s setter
		require.NoError(t, values.Set(&s, test.value), "%#v", test.value)
		assert.Equal(t, test.typ, s.typ, "%#v", test.value)
		assert.Equal(t, test.want, s.value, "%#v", test.value)
	}

	// Values which can't be, and a valuer's error
	failed := errors.New("failed")
	for _, value := range []any{uint64(1 << 63), []string{"a"}, struct{}{}, valuer{value: valuer{}}} {
		assert.Error(t, values.Set(&setter{}, value), "%#v", value)
	}
	assert.ErrorIs(t, values.Set(&setter{}, valuer{err: failed}), failed)
}

func Test_Scan(t *testing.T) {
	// Each kind, and named types
	var (
		b   bool
		i8  int8
		u16 uint16
		n   id
		f32 float32
		f64 float64
		s   string
		nm  name
		bs  []byte
		ts  time.Time
		a   any
		ns  sql.NullString
		ps  *string
	)
	for _, test := range []struct {
		value values.Value
		dest  any
		want  any
	}{
		{integer(1), &b, true},
		{text("false"), &b, false},
		{float(0.5), &b, true},
		{integer(-7), &i8, int8(-7)},
		{float(7), &u16, uint16(7)},
		{text(" 42 "), &n, id(42)},
		{integer(2), &f32, float32(2)},
		{text("2.5"), &f64, 2.5},
		{integer(42), &s, "42"},
		{text("b"), &nm, name("b")},
		{blob([]byte{1}), &bs, []byte{1}},
		{null(), &bs, []byte(nil)},
		{text("2026-11-18 12:30:00"), &ts, time.Date(2026, 11, 18, 12, 30, 0, 0, time.UTC)},
		{float(2.5), &a, 2.5},
		{null(), &a, nil},
		{text("x"), &ns, sql.NullString{String: "x", Valid: true}},
		{null(), &ns, sql.NullString{}},
		{null(), &ps, (*string)(nil)},
	} {
		require.NoError(t, values.Scan(test.value, test.dest), "%#v into %T", test.value, test.dest)
		got := fmt.Sprintf("%#v", deref(test.dest))
		assert.Equal(t, fmt.Sprintf("%#v", test.want), got, "%#v into %T", test.value, test.dest)
	}

	// A pointer, which isn't nil, of its value
	require.NoError(t, values.Scan(text("a"), &ps))
	if assert.NotNil(t, ps) {
		assert.Equal(t, "a", *ps)
	}

	// Values which don't fit, NULL, text which isn't, and not a pointer
	for _, test := range []struct {
		value values.Value
		dest  any
	}{
		{integer(300), &i8},
		{integer(-1), &u16},
		{float(1.5), &i8},
		{text("x"), &i8},
		{text("x"), &f64},
		{text("x"), &b},
		{blob([]byte{1}), &i8},
		{blob([]byte{1}), &f64},
		{blob([]byte{1}), &b},
		{null(), &s},
		{null(), &ts},
		{text("x"), &ts},
		{blob([]byte{1}), &ts},
		{integer(1), s},
		{integer(1), (*string)(nil)},
		{integer(1), &[]string{}},
	} {
		assert.Error(t, values.Scan(test.value, test.dest), "%#v into %T", test.value, test.dest)
	}
}

func Test_Time(t *testing.T) {
	want := time.Date(2026, 11, 18, 12, 30, 15, 0, time.UTC)
	for _, v := range []values.Value{
		text("2026-11-18T12:30:15Z"),
		text("2026-11-18T13:30:15+01:00"),
		text("2026-11-18 12:30:15"),
		text("2026-11-18T12:30:15.000"),
		integer(want.Unix()),
		float(2461363.0210069446),
	} {
		got, err := values.Time(v)
		require.NoError(t, err, "%#v", v)
		assert.WithinDuration(t, want, got, time.Millisecond, "%#v", v)
	}
	got, err := values.Time(text("2026-11-18"))
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC), got)
	_, err = values.Time(null())
	assert.Error(t, err)
}

func Test_Any(t *testing.T) {
	assert.Equal(t, int64(1), values.Any(integer(1)))
	assert.Equal(t, 1.5, values.Any(float(1.5)))
	assert.Equal(t, "a", values.Any(text("a")))
	assert.Equal(t, []byte{1}, values.Any(blob([]byte{1})))
	assert.Nil(t, values.Any(null()))
	assert.Equal(t, "INTEGER", values.Integer.String())
	assert.Equal(t, "UNKNOWN", values.Type(9).String())
}

// deref returns what the pointer points to
func deref(dest any) any {
	switch d := dest.(type) {
	case *bool:
		return *d
	case *int8:
		return *d
	case *uint16:
		return *d
	case *id:
		return *d
	case *float32:
		return *d
	case *float64:
		return *d
	case *string:
		return *d
	case *name:
		return *d
	case *[]byte:
		return *d
	case *time.Time:
		return *d
	case *any:
		return *d
	case *sql.NullString:
		return *d
	case **string:
		return *d
	}
	return strconv.Quote(fmt.Sprint(dest))
}
