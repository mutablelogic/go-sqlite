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
// it scans, and whose result is set as values.Set does
func createFunction(db *sqlite3.Conn, fn sqlite.Function) error {
	flag := sqlite3.FunctionFlag(0)
	if fn.Deterministic {
		flag |= sqlite3.DETERMINISTIC
	}
	return db.CreateFunction(fn.Name, fn.Args, flag, func(ctx sqlite3.Context, args ...sqlite3.Value) {
		in := make([]values.Value, len(args))
		for i, value := range args {
			in[i] = arg{value}
		}
		out, err := fn.Fn(in)
		if err == nil {
			err = values.Set(result{ctx}, out)
		}
		if err != nil {
			ctx.ResultError(fmt.Errorf("%s: %w", fn.Name, err))
		}
	})
}
