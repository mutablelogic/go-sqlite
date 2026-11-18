package sqlite

import (
	// Packages
	values "github.com/mutablelogic/go-sqlite/values"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Fn is a function, of its arguments, which SQL calls, which Func0, Func1,
// Func2, Func3 and FuncN make, of a Go function, of typed arguments and
// result
type Fn struct {
	args int
	fn   func(args []values.Value) (any, error)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Func0 returns a function, of no arguments, and its result, which is set as
// values.Set sets it, such as a []string, which is JSON
func Func0[T any](fn func() (T, error)) Fn {
	if fn == nil {
		return Fn{}
	}
	return Fn{args: 0, fn: func([]values.Value) (any, error) {
		return fn()
	}}
}

// Func1 returns a function, of one argument, which is scanned as values.Scan
// scans it, such as a time.Time, or a *int64, which is nil when it's NULL,
// and its result
func Func1[A, T any](fn func(A) (T, error)) Fn {
	if fn == nil {
		return Fn{}
	}
	return Fn{args: 1, fn: func(args []values.Value) (any, error) {
		var a A
		if err := scanArg(args, 0, &a); err != nil {
			return nil, err
		}
		return fn(a)
	}}
}

// Func2 returns a function, of two arguments, and its result
func Func2[A, B, T any](fn func(A, B) (T, error)) Fn {
	if fn == nil {
		return Fn{}
	}
	return Fn{args: 2, fn: func(args []values.Value) (any, error) {
		var a A
		var b B
		if err := scanArg(args, 0, &a); err != nil {
			return nil, err
		}
		if err := scanArg(args, 1, &b); err != nil {
			return nil, err
		}
		return fn(a, b)
	}}
}

// Func3 returns a function, of three arguments, and its result
func Func3[A, B, C, T any](fn func(A, B, C) (T, error)) Fn {
	if fn == nil {
		return Fn{}
	}
	return Fn{args: 3, fn: func(args []values.Value) (any, error) {
		var a A
		var b B
		var c C
		if err := scanArg(args, 0, &a); err != nil {
			return nil, err
		}
		if err := scanArg(args, 1, &b); err != nil {
			return nil, err
		}
		if err := scanArg(args, 2, &c); err != nil {
			return nil, err
		}
		return fn(a, b, c)
	}}
}

// FuncN returns a function, of any number of arguments, of one type, such
// as any, which is an int64, a float64, a string, a []byte or nil, and its
// result
func FuncN[A, T any](fn func(...A) (T, error)) Fn {
	if fn == nil {
		return Fn{}
	}
	return Fn{args: -1, fn: func(args []values.Value) (any, error) {
		in := make([]A, len(args))
		for i := range args {
			if err := scanArg(args, i, &in[i]); err != nil {
				return nil, err
			}
		}
		return fn(in...)
	}}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// scanArg scans an argument, which is an error, of its position, when it
// isn't of the type
func scanArg(args []values.Value, i int, dest any) error {
	if i >= len(args) {
		return ErrBadParameter.Withf("argument %d: missing", i+1)
	}
	if err := values.Scan(args[i], dest); err != nil {
		return ErrBadParameter.Withf("argument %d: %w", i+1, err)
	}
	return nil
}
