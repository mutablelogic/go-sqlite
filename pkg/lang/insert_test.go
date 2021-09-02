package lang_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_Insert_000(t *testing.T) {
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{N("foo").Insert(), `INSERT INTO foo DEFAULT VALUES`, ``},
		{N("foo").WithSchema("main").Insert(), `INSERT INTO main.foo DEFAULT VALUES`, ``},
		{N("foo").WithSchema("main").Insert("a"), `INSERT INTO main.foo (a) VALUES (?)`, ``},
		{N("foo").WithSchema("main").Insert("a", "b"), `INSERT INTO main.foo (a,b) VALUES (?,?)`, ``},
		{N("foo").Replace(), `REPLACE INTO foo DEFAULT VALUES`, ``},
		{N("foo").WithSchema("main").Replace(), `REPLACE INTO main.foo DEFAULT VALUES`, ``},
		{N("foo").WithSchema("main").Replace("a"), `REPLACE INTO main.foo (a) VALUES (?)`, ``},
		{N("foo").WithSchema("main").Replace("a", "b"), `REPLACE INTO main.foo (a,b) VALUES (?,?)`, ``},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("db.V = %v, wanted %v", v, test.String)
		}
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("db.V = %v, wanted %v", v, test.Query)
			}
		}
	}
}

func Test_Insert_001(t *testing.T) {
	tests := []struct {
		In    Statement
		Query string
	}{
		{N("foo").Insert("a").WithConflictDoNothing(), `INSERT INTO foo (a) VALUES (?) ON CONFLICT DO NOTHING`},
		{N("foo").Insert("a").WithConflictDoNothing().WithConflictDoNothing("a"), `INSERT INTO foo (a) VALUES (?) ON CONFLICT DO NOTHING ON CONFLICT (a) DO NOTHING`},
		{N("foo").Insert("a").WithConflictUpdate(), `INSERT INTO foo (a) VALUES (?) ON CONFLICT DO UPDATE SET a=excluded.a WHERE a<>excluded.a`},
		{N("foo").Insert("a").WithConflictUpdate("b"), `INSERT INTO foo (a) VALUES (?) ON CONFLICT (b) DO UPDATE SET a=excluded.a WHERE a<>excluded.a`},
		{N("foo").Insert("a").WithConflictUpdate("b", "c"), `INSERT INTO foo (a) VALUES (?) ON CONFLICT (b,c) DO UPDATE SET a=excluded.a WHERE a<>excluded.a`},
		{N("foo").Insert("a", "b").WithConflictUpdate(), `INSERT INTO foo (a,b) VALUES (?,?) ON CONFLICT DO UPDATE SET a=excluded.a,b=excluded.b WHERE a<>excluded.a OR b<>excluded.b`},
	}

	for _, test := range tests {
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("db.V = %v, wanted %v", v, test.Query)
			} else {
				t.Log(v)
			}
		}
	}
}
