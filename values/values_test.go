package values_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/netip"
	"strconv"
	"strings"
	"testing"
	"time"
	"uuid"

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
type octet uint8

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
		{time.Time{}, values.Null, nil},
		{&when, values.Text, "2026-11-18T12:30:00.000000123Z"},
		{&time.Time{}, values.Null, nil},
		{valuer{value: int64(3)}, values.Integer, int64(3)},
		{sql.NullString{}, values.Null, nil},

		// Durations are nanoseconds
		{90 * time.Minute, values.Integer, int64(90 * time.Minute)},

		// The standard library's UUIDs, which are text marshalers, are text
		{uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), values.Text, "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},

		// Text, binary and JSON, which marshal themselves, with text rather
		// than binary, when they're both
		{netip.MustParseAddr("10.0.0.1"), values.Text, "10.0.0.1"},
		{binaryOnly{1, 2}, values.Blob, []byte{1, 2}},
		{json.RawMessage(`{"a":1}`), values.Text, `{"a":1}`},

		// Composites are JSON, and arrays of bytes are blobs
		{[]string{"a", "b"}, values.Text, `["a","b"]`},
		{[]int{1, 2}, values.Text, `[1,2]`},
		{[2]string{"a", "b"}, values.Text, `["a","b"]`},
		{map[string]int{"a": 1}, values.Text, `{"a":1}`},
		{struct {
			A int `json:"a"`
		}{1}, values.Text, `{"a":1}`},
		{[3]byte{1, 2, 3}, values.Blob, []byte{1, 2, 3}},
		{[2]octet{1, 2}, values.Blob, []byte{1, 2}},
		{[]octet{1, 2}, values.Blob, []byte{1, 2}},
	} {
		var s setter
		require.NoError(t, values.Set(&s, test.value), "%#v", test.value)
		assert.Equal(t, test.typ, s.typ, "%#v", test.value)
		assert.Equal(t, test.want, s.value, "%#v", test.value)
	}

	// Values which can't be, and a valuer's error
	failed := errors.New("failed")
	for _, value := range []any{uint64(1 << 63), make(chan int), func() {}, []any{make(chan int)}, valuer{value: valuer{}}} {
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
		pt  *time.Time
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
		{null(), &ts, time.Time{}},
		{float(2.5), &a, 2.5},
		{null(), &a, nil},
		{text("x"), &ns, sql.NullString{String: "x", Valid: true}},
		{null(), &ns, sql.NullString{}},
		{null(), &ps, (*string)(nil)},
		{null(), &pt, (*time.Time)(nil)},
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

	// A Julian day which isn't one, or is past the year 9999
	for _, f := range []float64{1e300, math.Inf(1), math.NaN(), -1e300} {
		_, err := values.Time(float(f))
		assert.Error(t, err, f)
	}
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
	case **time.Time:
		return *d
	}
	return strconv.Quote(fmt.Sprint(dest))
}

func Test_ScanComposites(t *testing.T) {
	// The standard library's UUID, which unmarshals text, of text, or a
	// blob, of its bytes
	id := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var u uuid.UUID
	require.NoError(t, values.Scan(text(id.String()), &u))
	assert.Equal(t, id, u)
	u = uuid.UUID{}
	require.NoError(t, values.Scan(blob(id[:]), &u))
	assert.Equal(t, id, u)
	assert.Error(t, values.Scan(blob(id[:8]), &u))
	assert.Error(t, values.Scan(text("not a uuid"), &u))

	// Text, of a text unmarshaler, and a blob, of a binary one, which is
	// both, as netip.Addr is
	var addr netip.Addr
	require.NoError(t, values.Scan(text("10.0.0.1"), &addr))
	assert.Equal(t, netip.MustParseAddr("10.0.0.1"), addr)
	require.NoError(t, values.Scan(blob([]byte{10, 0, 0, 2}), &addr))
	assert.Equal(t, netip.MustParseAddr("10.0.0.2"), addr)
	var b binaryOnly
	require.NoError(t, values.Scan(blob([]byte{1, 2}), &b))
	assert.Equal(t, binaryOnly{1, 2}, b)
	assert.Error(t, values.Scan(text("x"), &addr))

	// JSON, of a json.Unmarshaler, and composites
	var raw json.RawMessage
	require.NoError(t, values.Scan(text(`{"a":1}`), &raw))
	assert.JSONEq(t, `{"a":1}`, string(raw))
	var strs []string
	require.NoError(t, values.Scan(text(`["a","b"]`), &strs))
	assert.Equal(t, []string{"a", "b"}, strs)
	var arr [2]int
	require.NoError(t, values.Scan(text(`[1,2]`), &arr))
	assert.Equal(t, [2]int{1, 2}, arr)
	var m map[string]int
	require.NoError(t, values.Scan(text(`{"a":1}`), &m))
	assert.Equal(t, map[string]int{"a": 1}, m)
	var st struct {
		A int `json:"a"`
	}
	require.NoError(t, values.Scan(text(`{"a":1}`), &st))
	assert.Equal(t, 1, st.A)

	// A map, or a struct, which was scanned into before, is of the value
	// alone, rather than merged into what it was
	m = map[string]int{"old": 1}
	require.NoError(t, values.Scan(text(`{"a":2}`), &m))
	assert.Equal(t, map[string]int{"a": 2}, m)
	st.A = 9
	require.NoError(t, values.Scan(text(`{}`), &st))
	assert.Zero(t, st.A)

	// As is one which unmarshals itself, of JSON, or text
	var labels jsonLabels
	require.NoError(t, values.Scan(text(`{"a":1}`), &labels))
	require.NoError(t, values.Scan(text(`{"b":2}`), &labels))
	assert.Equal(t, jsonLabels{"b": 2}, labels)
	var words textWords
	require.NoError(t, values.Scan(text("a b"), &words))
	require.NoError(t, values.Scan(text("c"), &words))
	assert.Equal(t, textWords{"c": true}, words)

	// NULL is a nil slice, or map, but not a struct
	require.NoError(t, values.Scan(null(), &strs))
	assert.Nil(t, strs)
	require.NoError(t, values.Scan(null(), &m))
	assert.Nil(t, m)
	assert.Error(t, values.Scan(null(), &st))

	// A blob is an array of bytes, of its length
	var bytes3 [3]byte
	require.NoError(t, values.Scan(blob([]byte{1, 2, 3}), &bytes3))
	assert.Equal(t, [3]byte{1, 2, 3}, bytes3)
	assert.Error(t, values.Scan(blob([]byte{1, 2}), &bytes3))
	assert.Error(t, values.Scan(text("abc"), &bytes3))

	// Arrays and slices of a named type of byte
	var octets [2]octet
	require.NoError(t, values.Scan(blob([]byte{1, 2}), &octets))
	assert.Equal(t, [2]octet{1, 2}, octets)
	var octetSlice []octet
	require.NoError(t, values.Scan(blob([]byte{1, 2}), &octetSlice))
	assert.Equal(t, []octet{1, 2}, octetSlice)
	assert.Error(t, values.Scan(blob([]byte{1}), &octets))

	// A nil scanner isn't a destination
	assert.Error(t, values.Scan(text("x"), (*sql.NullString)(nil)))

	// JSON which isn't, and which isn't text
	assert.Error(t, values.Scan(text(`not json`), &strs))
	assert.Error(t, values.Scan(blob([]byte(`["a"]`)), &strs))
	assert.Error(t, values.Scan(integer(1), &m))
}

func Test_Duration(t *testing.T) {
	for _, test := range []struct {
		value values.Value
		want  time.Duration
	}{
		{integer(int64(90 * time.Minute)), 90 * time.Minute},
		{float(1.5), 1500 * time.Millisecond},
		{text("1h30m"), 90 * time.Minute},
		{text("1000"), time.Microsecond},
	} {
		var d time.Duration
		require.NoError(t, values.Scan(test.value, &d), "%#v", test.value)
		assert.Equal(t, test.want, d, "%#v", test.value)
	}
	for _, v := range []values.Value{text("x"), blob([]byte{1}), float(1e300)} {
		_, err := values.Duration(v)
		assert.Error(t, err, "%#v", v)
	}
	var d time.Duration
	assert.Error(t, values.Scan(null(), &d))
}

// binaryOnly is a binary marshaler, which isn't a text one
type binaryOnly [2]byte

func (b binaryOnly) MarshalBinary() ([]byte, error) { return b[:], nil }
func (b *binaryOnly) UnmarshalBinary(data []byte) error {
	if len(data) != 2 {
		return errors.New("two bytes")
	}
	copy(b[:], data)
	return nil
}

// jsonLabels unmarshals itself, of JSON, as a map, which merges into what
// it was, unless it's empty
type jsonLabels map[string]int

func (l *jsonLabels) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, (*map[string]int)(l))
}

// textWords unmarshals itself, of text, which adds the words to what it was
type textWords map[string]bool

func (w *textWords) UnmarshalText(data []byte) error {
	if *w == nil {
		*w = make(textWords)
	}
	for _, word := range strings.Fields(string(data)) {
		(*w)[word] = true
	}
	return nil
}
