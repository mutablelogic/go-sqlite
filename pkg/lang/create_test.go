package lang_test

import (
	"fmt"
	"testing"

	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_Create_000(t *testing.T) {
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{N("foo").CreateTable(), `CREATE TABLE foo ()`, ``},
		{N("foo").CreateTable().WithTemporary(), `CREATE TEMPORARY TABLE foo ()`, ``},
		{N("foo").CreateTable().IfNotExists(), `CREATE TABLE IF NOT EXISTS foo ()`, ``},
		{N("foo").CreateTable(C("a"), C("b")), `CREATE TABLE foo (a TEXT,b TEXT)`, ``},
		{N("test").CreateTable(), "CREATE TABLE test ()", ""},
		{N("test").WithSchema("main").CreateTable(), "CREATE TABLE main.test ()", ""},
		{N("test").CreateTable().WithTemporary(), "CREATE TEMPORARY TABLE test ()", ""},
		{N("test").CreateTable().IfNotExists(), "CREATE TABLE IF NOT EXISTS test ()", ""},
		{N("test").CreateTable().WithoutRowID(), "CREATE TABLE test () WITHOUT ROWID", ""},
		{N("foo").CreateTable(C("a").NotNull(), C("b").NotNull()), `CREATE TABLE foo (a TEXT NOT NULL,b TEXT NOT NULL)`, ``},
		{N("foo").CreateTable(C("a").WithPrimary(), C("b")), `CREATE TABLE foo (a TEXT NOT NULL PRIMARY KEY,b TEXT)`, ``},
		{N("test").CreateTable(N("a").WithType("TEXT"), N("b").WithType("TEXT")), "CREATE TABLE test (a TEXT,b TEXT)", ""},
		{N("test").CreateTable(N("a").WithType("TEXT").NotNull(), N("b").WithType("TEXT").NotNull()), "CREATE TABLE test (a TEXT NOT NULL,b TEXT NOT NULL)", ""},
		{N("test").CreateTable(N("a").WithType("CUSTOM TYPE")), "CREATE TABLE test (a \"CUSTOM TYPE\")", ""},
		{N("test").CreateTable(N("a").WithType("CUSTOM TYPE").NotNull()), "CREATE TABLE test (a \"CUSTOM TYPE\" NOT NULL)", ""},
		{N("test").CreateTable(N("a").WithType("TEXT").WithPrimary()), "CREATE TABLE test (a TEXT NOT NULL PRIMARY KEY)", ""},
		{N("test").CreateTable(N("a").WithType("TEXT").WithPrimary(), N("b").WithType("TEXT").WithPrimary()), "CREATE TABLE test (a TEXT NOT NULL,b TEXT NOT NULL,PRIMARY KEY (a,b))", ""},
		{N("test").CreateTable(N("a").WithType("TEXT")).WithIndex("a"), "CREATE TABLE test (a TEXT,INDEX (a))", ""},
		{N("test").CreateTable(N("a").WithType("TEXT"), N("b").WithType("TEXT")).WithIndex("a", "b"), "CREATE TABLE test (a TEXT,b TEXT,INDEX (a,b))", ""},
		{N("test").CreateTable(N("a").WithType("TEXT"), N("b").WithType("TEXT")).WithUnique("a").WithUnique("b"), "CREATE TABLE test (a TEXT,b TEXT,UNIQUE (a),UNIQUE (b))", ""},
		{N("test").CreateTable(N("a").WithType("TEXT").WithPrimary(), N("b").WithType("TEXT")).WithUnique("b"), "CREATE TABLE test (a TEXT NOT NULL PRIMARY KEY,b TEXT,UNIQUE (b))", ""},
		{N("test").CreateTable(N("a").WithType("TEXT").WithAutoIncrement(), N("b").WithType("TEXT")).WithUnique("b"), "CREATE TABLE test (a TEXT NOT NULL PRIMARY KEY AUTOINCREMENT,b TEXT,UNIQUE (b))", ""},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("Unexpected return from String(): %q, wanted %q", v, test.String)
		} else {
			t.Log(v)
		}
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("Unexpected return from Query(): %q, wanted %q", v, test.Query)
			}
		}
	}
}

func Test_Create_001(t *testing.T) {
	tests := []struct {
		In     Statement
		String string
		Query  string
	}{
		{N("foo").CreateTable(N("a").WithType("TEXT")).WithForeignKey("a", N("bar").ForeignKey()), `CREATE TABLE foo (a TEXT,FOREIGN KEY (a) REFERENCES bar)`, ``},
		{N("foo").CreateTable(N("a").WithType("TEXT")).WithForeignKey("a", N("bar").ForeignKey("x", "y")), `CREATE TABLE foo (a TEXT,FOREIGN KEY (a) REFERENCES bar (x,y))`, ``},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("Unexpected return from String(): %q, wanted %q", v, test.String)
		} else {
			t.Log(v)
		}
		if test.Query != "" {
			if v := test.In.Query(); v != test.Query {
				t.Errorf("Unexpected return from Query(): %q, wanted %q", v, test.Query)
			}
		}
	}
}
