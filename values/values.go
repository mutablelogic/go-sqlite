// Package values converts values between Go and SQLite, of SQLite's five
// types: integers, floats, text, blobs and NULL. An implementation, such as
// package ncruces, adapts where SQLite gives values, such as a row's
// columns, to Value, and where it takes them, such as a statement's
// arguments, to Setter.
package values

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Type is one of SQLite's types of value
type Type int

// Value is a value which SQLite gives, such as a row's column, or a
// function's argument. Its methods convert it, as SQLite does, such as an
// integer to text.
type Value interface {
	Type() Type
	Int64() int64
	Float() float64
	Text() string
	Blob() []byte
}

// Setter is where SQLite takes a value, such as a statement's argument, or a
// function's result
type Setter interface {
	SetNull() error
	SetInt64(int64) error
	SetFloat(float64) error
	SetText(string) error
	SetBlob([]byte) error
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	Null Type = iota
	Integer
	Float
	Text
	Blob
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Type) String() string {
	switch t {
	case Null:
		return "NULL"
	case Integer:
		return "INTEGER"
	case Float:
		return "FLOAT"
	case Text:
		return "TEXT"
	case Blob:
		return "BLOB"
	}
	return "UNKNOWN"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Any returns the value as an int64, a float64, a string, a []byte, or nil
func Any(v Value) any {
	switch v.Type() {
	case Integer:
		return v.Int64()
	case Float:
		return v.Float()
	case Text:
		return v.Text()
	case Blob:
		return v.Blob()
	}
	return nil
}
