package sqlite_test

import (
	"sync"
	"testing"

	// Packages
	sqlite "github.com/mutablelogic/go-sqlite"
	"github.com/stretchr/testify/assert"
)

func Test_Bind_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		bind := sqlite.NewBind("a", "b")
		assert.NotNil(bind)
		assert.True(bind.Has("a"))
		assert.Equal("b", bind.Get("a"))
	})

	t.Run("2", func(t *testing.T) {
		bind := sqlite.NewBind("a", "b", "c")
		assert.Nil(bind)
	})

	t.Run("3", func(t *testing.T) {
		bind := sqlite.NewBind("a", 100)
		assert.NotNil(bind)
		assert.True(bind.Has("a"))
		assert.Equal(100, bind.Get("a"))
	})

	t.Run("4", func(t *testing.T) {
		bind := sqlite.NewBind()
		assert.NotNil(bind)
		assert.Equal("@a", bind.Set("a", "b"))
		assert.True(bind.Has("a"))
		assert.Equal("b", bind.Get("a"))
	})

	t.Run("5", func(t *testing.T) {
		bind := sqlite.NewBind("", "b")
		assert.Nil(bind)
	})

	t.Run("6", func(t *testing.T) {
		bind := sqlite.NewBind()
		assert.NotNil(bind)
		assert.Equal("", bind.Set("", "b"))
	})

}

func Test_Bind_002(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		In  string
		Out string
	}{
		{In: `$schema`, Out: "$schema"},
		{In: `VALUES ($schema, @schema, :schema)`, Out: "VALUES ($schema, @schema, :schema)"},
		{In: `${schema} ${schema`, Out: "schema ${schema"},
		{In: `${'schema'}`, Out: "'schema'"},
		{In: `${"schema"}`, Out: `"schema"`},
		{In: `$1`, Out: `$1`},
		{In: `${1}`, Out: `$1`},
		{In: `$$`, Out: `$$`},
		{In: `${'single'}`, Out: `'''single'''`},
		{In: `${"single"}`, Out: `"'single'"`},
		{In: `${'double'}`, Out: `'"double"'`},
		{In: `${"double"}`, Out: `"""double"""`},
	}

	bind := sqlite.NewBind(
		"schema", "schema",
		"single", "'single'",
		"double", "\"double\"",
	)

	for _, test := range tests {
		t.Run(test.In, func(t *testing.T) {
			assert.Equal(test.Out, bind.Replace(test.In))
		})
	}
}

func Test_Bind_003(t *testing.T) {
	assert := assert.New(t)

	bind := sqlite.NewBind(
		"list", []string{"a", "b", "c"},
	)
	assert.Equal("IN ('a','b','c')", bind.Replace("IN (${'list'})"))
}

func Test_Query_001(t *testing.T) {
	assert := assert.New(t)

	// Test Bind.Query method - sets span name and returns the named SQL
	bind := sqlite.NewBind("user.select", "SELECT * FROM users WHERE id = @id")
	sql := bind.Query("user.select")
	assert.Equal("SELECT * FROM users WHERE id = @id", sql)
	assert.Equal("user.select", bind.Get(sqlite.TraceSpanNameArg))
	assert.Empty(bind.Query("user.missing"))
}

func Test_Query_002(t *testing.T) {
	assert := assert.New(t)

	// The named SQL's substitutions are replaced once, when it's executed,
	// so a value with a $ in it isn't replaced again
	bind := sqlite.NewBind("file.get", `SELECT * FROM ${"table"} WHERE path = ${'path'}`, "table", "files", "path", "a$b.txt", "b", "oops")
	assert.Equal(`SELECT * FROM "files" WHERE path = 'a$b.txt'`, bind.Replace(bind.Query("file.get")))
}

func Test_Bind_Concurrent(t *testing.T) {
	// A bind is read, replaced and marshaled while it's set, without a race
	bind := sqlite.NewBind("q", "SELECT ${a}")
	var wg sync.WaitGroup
	for i := range 4 {
		wg.Go(func() {
			for j := range 100 {
				bind.Set("a", i*j)
				_ = bind.Replace(bind.Query("q"))
				_, _ = bind.MarshalJSON()
				_ = bind.String()
				_ = bind.Copy()
			}
		})
	}
	wg.Wait()
}

func Test_Bind_Copy_001(t *testing.T) {
	assert := assert.New(t)

	bind := sqlite.NewBind("a", "b")
	copy := bind.Copy()
	if assert.NotNil(copy) {
		assert.Equal("b", copy.Get("a"))
	}

	copy.Set("a", "c")
	assert.Equal("b", bind.Get("a"))
	assert.Equal("c", copy.Get("a"))
}

func Test_Bind_Copy_002(t *testing.T) {
	assert := assert.New(t)

	bind := sqlite.NewBind("a", "b")
	copy := bind.Copy("c", "d")
	if assert.NotNil(copy) {
		assert.Equal("b", copy.Get("a"))
		assert.Equal("d", copy.Get("c"))
	}
	assert.False(bind.Has("c"))
}
