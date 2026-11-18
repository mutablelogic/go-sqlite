package sqlite

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"strings"
	"sync"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Bind represents a set of variables and arguments to be used in a query.
// The vars are substituted in the query string itself, as ${key}, while the
// args are passed as arguments to the query, as @key.
type Bind struct {
	sync.RWMutex
	vars map[string]any
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// TraceSpanNameArg is the bind var which names a query's span
	TraceSpanNameArg = "otelspan"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewBind creates a new Bind object with the given name/value pairs.
// Returns nil if the number of arguments is not even.
func NewBind(pairs ...any) *Bind {
	if len(pairs)%2 != 0 {
		return nil
	}

	// Populate the vars map
	vars := make(map[string]any, len(pairs)>>1)
	for i := 0; i < len(pairs); i += 2 {
		if key, ok := pairs[i].(string); !ok || key == "" {
			return nil
		} else {
			vars[key] = pairs[i+1]
		}
	}

	// Return the Bind object
	return &Bind{vars: vars}
}

// Copy creates a copy of the bind object with additional name/value pairs.
func (bind *Bind) Copy(pairs ...any) *Bind {
	if len(pairs)%2 != 0 {
		return nil
	}
	if len(pairs) == 0 {
		bind.RLock()
		defer bind.RUnlock()
		return &Bind{vars: maps.Clone(bind.vars)}
	}

	// Lock before copying
	varsCopy := func() map[string]any {
		bind.RLock()
		defer bind.RUnlock()
		c := make(map[string]any, len(bind.vars)+(len(pairs)>>1))
		maps.Copy(c, bind.vars)
		return c
	}()

	for i := 0; i < len(pairs); i += 2 {
		if key, ok := pairs[i].(string); !ok || key == "" {
			return nil
		} else {
			varsCopy[key] = pairs[i+1]
		}
	}

	// Return the copied Bind object
	return &Bind{vars: varsCopy}
}

// Return a new bind object with one or more sets of queries
func (bind *Bind) withQueries(queries ...*Queries) *Bind {
	if len(queries) == 0 {
		return bind
	}

	// Make a copy of the bind vars
	varsCopy := bind.snapshot()

	// Iterate through queries
	for _, q := range queries {
		for _, key := range q.Keys() {
			varsCopy[key] = q.Query(key)
		}
	}

	// Return the copied Bind object
	return &Bind{vars: varsCopy}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (bind *Bind) MarshalJSON() ([]byte, error) {
	return json.Marshal(bind.snapshot())
}

func (bind *Bind) String() string {
	data, err := json.MarshalIndent(bind.snapshot(), "", "  ")
	if err != nil {
		return err.Error()
	} else {
		return string(data)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set sets a bind var and returns the parameter name.
func (bind *Bind) Set(key string, value any) string {
	bind.Lock()
	defer bind.Unlock()

	if key == "" {
		return ""
	}
	bind.vars[key] = value
	return "@" + key
}

// Get returns a bind var by key.
func (bind *Bind) Get(key string) any {
	bind.RLock()
	defer bind.RUnlock()
	return bind.vars[key]
}

// Has returns true if there is a bind var with the given key.
func (bind *Bind) Has(key string) bool {
	bind.RLock()
	defer bind.RUnlock()

	_, ok := bind.vars[key]
	return ok
}

// Del deletes a bind var.
func (bind *Bind) Del(key string) {
	bind.Lock()
	defer bind.Unlock()
	delete(bind.vars, key)
}

// Join joins a bind var with a separator when it is a
// []any and returns the result as a string. Returns
// an empty string if the key does not exist.
func (bind *Bind) Join(key, sep string) string {
	bind.RLock()
	defer bind.RUnlock()

	if value, ok := bind.vars[key]; !ok {
		return ""
	} else if v, ok := value.([]any); ok {
		str := make([]string, len(v))
		for i, value := range v {
			str[i] = fmt.Sprint(value)
		}
		return strings.Join(str, sep)
	} else {
		return fmt.Sprint(value)
	}
}

// Append appends a bind var to a list. Returns false if the key
// is not a list, or the value is not a list.
func (bind *Bind) Append(key string, value any) bool {
	bind.Lock()
	defer bind.Unlock()

	// Create a new list if it doesn't exist
	if _, ok := bind.vars[key]; !ok {
		bind.vars[key] = make([]any, 0, 5)
	}

	// Check type
	if _, ok := bind.vars[key].([]any); !ok {
		return false
	}

	// Append value
	bind.vars[key] = append(bind.vars[key].([]any), value)

	// Return success
	return true
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

// Query looks up a named query, which was loaded via WithQueries, such as
// "index.object.get", sets the span name for OpenTelemetry tracing, and
// returns its SQL, whose ${key} substitutions are replaced once, when it's
// executed, so a value which has a $ in it, such as a path, isn't replaced
// again. It returns an empty string when there's no query of the name.
func (bind *Bind) Query(name string) string {
	bind.Set(TraceSpanNameArg, name)
	sql, _ := bind.Get(name).(string)
	return sql
}

// Replace returns a query string with ${subtitution} replaced by the values:
//   - ${key} => value
//   - ${'key'} => 'value'
//   - ${"key"} => "value"
//   - $1 => $1
//   - $$ => $$
func (bind *Bind) Replace(query string) string {
	return replace(query, bind.snapshot())
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// snapshot returns a copy of the bind vars, so they're read, formatted or
// marshaled without the lock held, while they may be set
func (bind *Bind) snapshot() map[string]any {
	bind.RLock()
	defer bind.RUnlock()
	return maps.Clone(bind.vars)
}

func replace(query string, vars map[string]any) string {
	fetch := func(key string) string {
		return fmt.Sprint(vars[key])
	}
	return os.Expand(query, func(key string) string {
		if key == "$" { // $$ => $$
			return "$$"
		}
		if types.IsNumeric(key) {
			return "$" + key // $1 => $1
		}
		if types.IsSingleQuoted(key) { // ${'key'} => 'value'
			// Special case where value is []string and single quote for IN (${key})
			key := strings.Trim(key, "'")
			value := vars[key]
			switch v := value.(type) {
			case []string:
				result := make([]string, len(v))
				for i, s := range v {
					result[i] = types.Quote(s)
				}
				return strings.Join(result, ",")
			default:
				return types.Quote(fetch(key))
			}
		}
		if types.IsDoubleQuoted(key) { // ${"key"} => "value"
			return types.DoubleQuote(fetch(strings.Trim(key, "\"")))
		}
		return fetch(key) // ${key} => value
	})
}
