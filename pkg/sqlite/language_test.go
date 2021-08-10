package sqlite_test

import (
	"fmt"
	"testing"

	sq "github.com/djthorpe/go-sqlite"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
)

func Test_Lang_001(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check db.P()
	param := db.P()
	if v := fmt.Sprint(param); v != "?" {
		t.Errorf("db.P() = %v, wanted ?", v)
	}
	if v := param.Query(); v != "SELECT ?" {
		t.Errorf("db.P() = %v, wanted SELECT ?", v)
	}
}

func Test_Lang_002(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check db.V
	tests := []struct {
		In     sq.SQExpr
		String string
		Query  string
	}{
		{db.V(nil), "NULL", "SELECT NULL"},
		{db.V(1), "1", "SELECT 1"},
		{db.V(1.1), "1.1", "SELECT 1.1"},
		{db.V(true), "TRUE", "SELECT TRUE"},
		{db.V(false), "FALSE", "SELECT FALSE"},
		{db.V("foo"), "'foo'", "SELECT 'foo'"},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("db.V = %v, wanted %v", v, test.String)
		}
		if v := test.In.Query(); v != test.Query {
			t.Errorf("db.V = %v, wanted %v", v, test.Query)
		}
	}
}

func Test_Lang_003(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check db.Q
	tests := []struct {
		In    sq.SQStatement
		Query string
	}{
		{db.Q(nil), "SELECT NULL"},
		{db.Q(1), "SELECT 1"},
		{db.Q(1.1), "SELECT 1.1"},
		{db.Q(true), "SELECT TRUE"},
		{db.Q(false), "SELECT FALSE"},
		{db.Q("SELECT * FROM foo"), "SELECT * FROM foo"},
		{db.Q(""), "SELECT ''"},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("db.Q = %v, wanted %v", v, test.Query)
		}
	}
}
func Test_Lang_004(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check db.N
	tests := []struct {
		In     sq.SQName
		String string
		Query  string
	}{
		{db.N("foo"), "foo", "SELECT * FROM foo"},
		{db.N("insert"), "\"insert\"", "SELECT * FROM \"insert\""},
		{db.N("two words"), "\"two words\"", "SELECT * FROM \"two words\""},
		{db.N("two words").WithAlias("alias"), "\"two words\" AS alias", "SELECT * FROM \"two words\" AS alias"},
		{db.N("two words").WithSchema("main"), "main.\"two words\"", "SELECT * FROM main.\"two words\""},
		{db.N("foo").WithAlias("alias"), "foo AS alias", "SELECT * FROM foo AS alias"},
		{db.N("foo").WithSchema("main").WithAlias("alias"), "main.foo AS alias", "SELECT * FROM main.foo AS alias"},
	}

	for _, test := range tests {
		if v := fmt.Sprint(test.In); v != test.String {
			t.Errorf("db.N = %v, wanted %v", v, test.String)
		}
		if v := test.In.Query(); v != test.Query {
			t.Errorf("db.N = %v, wanted %v", v, test.Query)
		}
	}
}

func Test_Lang_005(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check db.N.CreateTable
	tests := []struct {
		In    sq.SQTable
		Query string
	}{
		{db.N("test").CreateTable(), "CREATE TABLE test ()"},
		{db.N("test").WithSchema("main").CreateTable(), "CREATE TABLE main.test ()"},
		{db.N("test").CreateTable().WithTemporary(), "CREATE TEMPORARY TABLE test ()"},
		{db.N("test").CreateTable().IfNotExists(), "CREATE TABLE IF NOT EXISTS test ()"},
		{db.N("test").CreateTable().WithoutRowID(), "CREATE TABLE test () WITHOUT ROWID"},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("db.N.CreateTable = %v, wanted %v", v, test.Query)
		}
	}
}

func Test_Lang_006(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check db.N.CreateTable
	tests := []struct {
		In    sq.SQTable
		Query string
	}{
		{db.N("test").CreateTable(db.N("a").WithType("TEXT"), db.N("b").WithType("TEXT")), "CREATE TABLE test (a TEXT,b TEXT)"},
		{db.N("test").CreateTable(db.N("a").WithType("TEXT").NotNull(), db.N("b").WithType("TEXT").NotNull()), "CREATE TABLE test (a TEXT NOT NULL,b TEXT NOT NULL)"},
		{db.N("test").CreateTable(db.N("a").WithType("CUSTOM TYPE")), "CREATE TABLE test (a \"CUSTOM TYPE\")"},
		{db.N("test").CreateTable(db.N("a").WithType("CUSTOM TYPE").NotNull()), "CREATE TABLE test (a \"CUSTOM TYPE\" NOT NULL)"},
		{db.N("test").CreateTable(db.N("a").WithType("TEXT").Primary()), "CREATE TABLE test (a TEXT NOT NULL,PRIMARY KEY (a))"},
		{db.N("test").CreateTable(db.N("a").WithType("TEXT").Primary(), db.N("b").WithType("TEXT").Primary()), "CREATE TABLE test (a TEXT NOT NULL,b TEXT NOT NULL,PRIMARY KEY (a,b))"},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("db.N.CreateTable = %v, wanted %v", v, test.Query)
		}
	}
}
func Test_Lang_007(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check db.N.CreateTable
	tests := []struct {
		In    sq.SQTable
		Query string
	}{
		{db.N("test").CreateTable(db.N("a").WithType("TEXT")).WithIndex("a"), "CREATE TABLE test (a TEXT,INDEX (a))"},
		{db.N("test").CreateTable(db.N("a").WithType("TEXT"), db.N("b").WithType("TEXT")).WithIndex("a", "b"), "CREATE TABLE test (a TEXT,b TEXT,INDEX (a,b))"},
		{db.N("test").CreateTable(db.N("a").WithType("TEXT"), db.N("b").WithType("TEXT")).WithUnique("a").WithUnique("b"), "CREATE TABLE test (a TEXT,b TEXT,UNIQUE (a),UNIQUE (b))"},
		{db.N("test").CreateTable(db.N("a").WithType("TEXT").Primary(), db.N("b").WithType("TEXT")).WithUnique("b"), "CREATE TABLE test (a TEXT NOT NULL,b TEXT,PRIMARY KEY (a),UNIQUE (b))"},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("db.N.CreateTable = %v, wanted %v", v, test.Query)
		}
	}
}

func Test_Lang_008(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check db.N.DropTable
	tests := []struct {
		In    sq.SQDrop
		Query string
	}{
		{db.N("test").DropTable(), "DROP TABLE test"},
		{db.N("test").DropTable().IfExists(), "DROP TABLE IF EXISTS test"},
		{db.N("test").WithSchema("main").DropTable().IfExists(), "DROP TABLE IF EXISTS main.test"},
		{db.N("test").WithSchema("main").DropView().IfExists(), "DROP VIEW IF EXISTS main.test"},
		{db.N("test").WithSchema("main").DropTrigger().IfExists(), "DROP TRIGGER IF EXISTS main.test"},
		{db.N("test").WithSchema("main").DropIndex().IfExists(), "DROP INDEX IF EXISTS main.test"},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("db.N.DropTable = %v, wanted %v", v, test.Query)
		}
	}
}
func Test_Lang_009(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check db.S (select)
	tests := []struct {
		In    sq.SQStatement
		Query string
	}{
		{db.S(), "SELECT *"},
		{db.S(db.N("a")), "SELECT * FROM a"},
		{db.S(db.N("a").WithAlias("aa")), "SELECT * FROM a AS aa"},
		{db.S(db.N("a"), db.N("b")), "SELECT * FROM a,b"},
		{db.S(db.N("a").WithAlias("aa"), db.N("b").WithAlias("bb")), "SELECT * FROM a AS aa,b AS bb"},
		{db.S().WithDistinct(), "SELECT DISTINCT *"},
		{db.S().WithLimitOffset(0, 0), "SELECT *"},
		{db.S().WithLimitOffset(1, 0), "SELECT * LIMIT 1"},
		{db.S(db.N("a")).WithLimitOffset(1, 0), "SELECT * FROM a LIMIT 1"},
		{db.S(db.N("a")).WithLimitOffset(0, 1), "SELECT * FROM a OFFSET 1"},
		{db.S(db.N("a")).WithLimitOffset(1, 1), "SELECT * FROM a LIMIT 1,1"},
		{db.S().Where(nil), "SELECT * WHERE NULL"},
		{db.S(db.N("a")).Where(nil), "SELECT * FROM a WHERE NULL"},
		{db.S(db.N("a")).Where(nil, nil), "SELECT * FROM a WHERE NULL AND NULL"},
		{db.S(db.N("a")).Where(db.N("a")), "SELECT * FROM a WHERE a"},
		{db.S(db.N("a")).Where(db.N("b"), db.N("c")), "SELECT * FROM a WHERE b AND c"},
		{db.S(db.N("a")).Where(db.P(), db.P()), "SELECT * FROM a WHERE ? AND ?"},
		{db.S(db.N("a")).Where(db.V("foo"), db.V(true)), "SELECT * FROM a WHERE 'foo' AND TRUE"},
		{db.S(db.N("a")).Where(db.V("foo"), db.V(false)), "SELECT * FROM a WHERE 'foo' AND FALSE"},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("db.S = %v, wanted %v", v, test.Query)
		}
	}
}

func Test_Lang_010(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check db.N.CreateView (select)
	tests := []struct {
		In    sq.SQStatement
		Query string
	}{
		{db.N("foo").CreateView(db.S()), "CREATE VIEW foo AS SELECT *"},
		{db.N("foo").WithSchema("main").CreateView(db.S()), "CREATE VIEW main.foo AS SELECT *"},
		{db.N("foo").CreateView(db.S()).IfNotExists(), "CREATE VIEW IF NOT EXISTS foo AS SELECT *"},
		{db.N("foo").CreateView(db.S()).WithTemporary(), "CREATE TEMPORARY VIEW foo AS SELECT *"},
		{db.N("foo").CreateView(db.S(), "a", "b"), "CREATE VIEW foo (a,b) AS SELECT *"},
	}

	for _, test := range tests {
		if v := test.In.Query(); v != test.Query {
			t.Errorf("db.N.CreateView = %v, wanted %v", v, test.Query)
		}
	}
}

/*

func Test_Lang_007(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if st := db.CreateIndex("test", "a", "b"); st.Query() != "CREATE INDEX test_a_b ON test (a,b)" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.CreateIndex("test", "a", "b").IfNotExists(); st.Query() != "CREATE INDEX IF NOT EXISTS test_a_b ON test (a,b)" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.CreateIndex("test", "a", "b").WithUnique().IfNotExists(); st.Query() != "CREATE UNIQUE INDEX IF NOT EXISTS test_a_b ON test (a,b)" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.CreateIndex("test", "a", "b").WithUnique().WithSchema("main"); st.Query() != "CREATE UNIQUE INDEX main.test_a_b ON test (a,b)" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.DropIndex("test"); st.Query() != "DROP INDEX test" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.DropIndex("test").IfExists(); st.Query() != "DROP INDEX IF EXISTS test" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.DropIndex("test").IfExists().WithSchema("main"); st.Query() != "DROP INDEX IF EXISTS main.test" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
}

func Test_Lang_009(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(db.CreateTable("test", db.Column("a", "TEXT"))); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(db.CreateIndex("test", "a").WithName("test_idx")); err != nil {
		t.Fatal(err)
	}
}
func Test_Lang_010(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if st := db.Insert("test"); st.Query() != "INSERT INTO test DEFAULT VALUES" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Insert("test").WithSchema("main"); st.Query() != "INSERT INTO main.test DEFAULT VALUES" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Insert("test", "a"); st.Query() != "INSERT INTO test (a) VALUES (?)" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Insert("test", "a", "b"); st.Query() != "INSERT INTO test (a,b) VALUES (?,?)" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
}

func Test_Lang_011(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if st := db.Select(); st.Query() != "SELECT *" {
		t.Errorf("Unexpected return: %q", st.Query())
	}

	if st := db.Select().WithDistinct(); st.Query() != "SELECT DISTINCT *" {
		t.Errorf("Unexpected return: %q", st.Query())
	}

	if st := db.Select().WithLimitOffset(0, 0); st.Query() != "SELECT *" {
		t.Errorf("Unexpected return: %q", st.Query())
	}

	if st := db.Select().WithLimitOffset(0, 1); st.Query() != "SELECT * OFFSET 1" {
		t.Errorf("Unexpected return: %q", st.Query())
	}

	if st := db.Select().WithLimitOffset(100, 0); st.Query() != "SELECT * LIMIT 100" {
		t.Errorf("Unexpected return: %q", st.Query())
	}

	if st := db.Select().WithLimitOffset(100, 1); st.Query() != "SELECT * LIMIT 100,1" {
		t.Errorf("Unexpected return: %q", st.Query())
	}

	if st := db.Select(db.TableSource("test")); st.Query() != "SELECT * FROM test" {
		t.Errorf("Unexpected return: %q", st.Query())
	}

	if st := db.Select(db.TableSource("a"), db.TableSource("b")); st.Query() != "SELECT * FROM a,b" {
		t.Errorf("Unexpected return: %q", st.Query())
	}

	if st := db.Select(db.TableSource("a").WithAlias("alias_a"), db.TableSource("b").WithAlias("alias_b")); st.Query() != "SELECT * FROM a AS alias_a,b AS alias_b" {
		t.Errorf("Unexpected return: %q", st.Query())
	}

}

func Test_Lang_012(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if st := db.Select().AsView("foo"); st.Query() != "CREATE VIEW foo AS SELECT *" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Select().AsView("foo").WithTemporary(); st.Query() != "CREATE TEMPORARY VIEW foo AS SELECT *" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Select().AsView("foo").IfNotExists(); st.Query() != "CREATE VIEW IF NOT EXISTS foo AS SELECT *" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Select().AsView("foo").WithSchema("main"); st.Query() != "CREATE VIEW main.foo AS SELECT *" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Select().AsView("foo", "a").WithSchema("main"); st.Query() != "CREATE VIEW main.foo (a) AS SELECT *" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Select().AsView("foo", "a", "b").WithSchema("main"); st.Query() != "CREATE VIEW main.foo (a,b) AS SELECT *" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
}
func Test_Lang_013(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if st := db.Q(nil); st.Query() != "SELECT NULL" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Select().Where(nil); st.Query() != "SELECT * WHERE NULL" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Select().Where("a").Where("b"); st.Query() != "SELECT * WHERE 'a' AND 'b'" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Select().Where(100).Where(200); st.Query() != "SELECT * WHERE 100 AND 200" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Select().Where(true).Where(false); st.Query() != "SELECT * WHERE TRUE AND FALSE" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Select().Where(time.Time{}); st.Query() != "SELECT * WHERE NULL" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	now := time.Now()
	if st := db.Select().Where(now); st.Query() != "SELECT * WHERE "+sqlite.Quote(now.Format(time.RFC3339Nano)) {
		t.Errorf("Unexpected return: %q", st.Query())
	}
}

func Test_Lang_014(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if st := db.Is(db.V("a"), nil); st.Query() != "SELECT a IS NULL" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Is(db.V("a"), nil).Not(); st.Query() != "SELECT a IS NOT NULL" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Is(db.V("a"), nil, nil); st.Query() != "SELECT (a IS NULL OR a IS NULL)" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.Is(db.V("a"), nil, nil).Not(); st.Query() != "SELECT NOT(a IS NULL OR a IS NULL)" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
}

func Test_Lang_015(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if st := db.V(""); st.Query() != "SELECT NULL" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.P(); st.Query() != "SELECT ?" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.V("a"); st.Query() != "SELECT a" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.V("insert"); st.Query() != "SELECT \"insert\"" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
}
*/
