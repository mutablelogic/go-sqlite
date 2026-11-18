package ncruces

import (
	"fmt"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	values "github.com/mutablelogic/go-sqlite/values"
	sqlite3 "github.com/ncruces/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// createFunction creates the function on the connection, whose arguments
// are int64, float64, string, []byte or nil, and whose result is set as
// values.Set does
func createFunction(db *sqlite3.Conn, fn sqlite.Function) error {
	flag := sqlite3.FunctionFlag(0)
	if fn.Deterministic {
		flag |= sqlite3.DETERMINISTIC
	}
	return db.CreateFunction(fn.Name, fn.Args, flag, func(ctx sqlite3.Context, args ...sqlite3.Value) {
		in := make([]any, len(args))
		for i, value := range args {
			in[i] = values.Any(arg{value})
		}
		out, err := fn.Fn(in...)
		if err == nil {
			err = values.Set(result{ctx}, out)
		}
		if err != nil {
			ctx.ResultError(fmt.Errorf("%s: %w", fn.Name, err))
		}
	})
}
