package sqlite_test

import (
	"strconv"
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

	tests := []struct {
		f     func() sq.SQTable
		query string
	}{
		{func() sq.SQTable { return db.CreateTable("test") }, "CREATE TABLE test ()"},
		{func() sq.SQTable { return db.CreateTable("test").WithSchema("test") }, "CREATE TABLE test.test ()"},
		{func() sq.SQTable { return db.CreateTable("test").WithTemporary() }, "CREATE TEMPORARY TABLE test ()"},
		{func() sq.SQTable { return db.CreateTable("test").IfNotExists() }, "CREATE TABLE IF NOT EXISTS test ()"},
		{func() sq.SQTable { return db.CreateTable("test").WithoutRowID() }, "CREATE TABLE test () WITHOUT ROWID"},
	}

	for i, test := range tests {
		if statement := test.f(); statement == nil {
			t.Errorf("Test %v: nil value returned", i)
		} else if statement.Query() != test.query {
			t.Errorf("Test %v: Expected %v, got %v", i, strconv.Quote(test.query), strconv.Quote(statement.Query()))
		}
	}
}

func Test_Lang_002(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if statement := db.CreateTable("test", db.Column("a", "TEXT").NotNull(), db.Column("b", "TEXT")); statement == nil {
		t.Error("Statement returned is nil")
	} else if statement.Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT)" {
		t.Errorf("Unexpected value, %v", statement.Query())
	}
}

func Test_Lang_003(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if statement := db.CreateTable("test", db.Column("a", "TEXT").Primary(), db.Column("b", "TEXT")); statement == nil {
		t.Error("Statement returned is nil")
	} else if statement.Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT,PRIMARY KEY (a))" {
		t.Errorf("Unexpected value, %v", statement.Query())
	}
}

func Test_Lang_004(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if statement := db.CreateTable("test", db.Column("a", "TEXT").Primary(), db.Column("b", "TEXT")); statement == nil {
		t.Error("Statement returned is nil")
	} else if statement.WithUnique("a", "b").Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT,PRIMARY KEY (a),UNIQUE (a,b))" {
		t.Errorf("Unexpected value, %v", statement.Query())
	}

	if statement := db.CreateTable("test", db.Column("a", "TEXT").NotNull(), db.Column("b", "TEXT")); statement == nil {
		t.Error("Statement returned is nil")
	} else if statement.WithUnique("a").WithUnique("b").Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT,UNIQUE (a),UNIQUE (b))" {
		t.Errorf("Unexpected value, %v", statement.Query())
	}

	if statement := db.CreateTable("test", db.Column("a", "TEXT").Primary(), db.Column("b", "TEXT").Primary()); statement == nil {
		t.Error("Statement returned is nil")
	} else if statement.Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT NOT NULL,PRIMARY KEY (a,b))" {
		t.Errorf("Unexpected value, %v", statement.Query())
	}

	if statement := db.CreateTable("test", db.Column("a", "TEXT").Primary(), db.Column("b", "TEXT").Primary()); statement == nil {
		t.Error("Statement returned is nil")
	} else if statement.WithUnique("a", "b").Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT NOT NULL,PRIMARY KEY (a,b),UNIQUE (a,b))" {
		t.Errorf("Unexpected value, %v", statement.Query())
	}

	if statement := db.CreateTable("test", db.Column("a", "TEXT").Primary(), db.Column("b", "TEXT").Primary()); statement == nil {
		t.Error("Statement returned is nil")
	} else if statement.WithIndex("a", "b").Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT NOT NULL,PRIMARY KEY (a,b),INDEX (a,b))" {
		t.Errorf("Unexpected value, %v", statement.Query())
	}
}

func Test_Lang_005(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	statement := db.CreateTable("test", db.Column("a", "TEXT").Primary(), db.Column("b", "TEXT"))
	if _, err := db.Exec(statement); err != nil {
		t.Fatal(err)
	} else if n := db.Tables(); n == nil {
		t.Error("Unexpected nil return from Tables()")
	} else if len(n) != 1 {
		t.Error("Unexpected number of tables", n)
	} else if n[0] != "test" {
		t.Errorf("Unexpected table name %q", n[0])
	} else if columns := db.Columns("test"); columns == nil {
		t.Error("Unexpected nil return from Columns()")
	} else if len(columns) != 2 {
		t.Error("Unexpected number of columns", len(columns))
	} else {
		t.Log(columns)
	}
}

func Test_Lang_006(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if st := db.DropTable("test"); st.Query() != "DROP TABLE test" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.DropTable("test").WithSchema("main"); st.Query() != "DROP TABLE main.test" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.DropTable("test").WithSchema("main").IfExists(); st.Query() != "DROP TABLE IF EXISTS main.test" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
}

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

func Test_Lang_008(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if st := db.DropIndex("test"); st.Query() != "DROP INDEX test" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.DropIndex("test").WithSchema("main"); st.Query() != "DROP INDEX main.test" {
		t.Errorf("Unexpected return: %q", st.Query())
	}
	if st := db.DropIndex("test").WithSchema("main").IfExists(); st.Query() != "DROP INDEX IF EXISTS main.test" {
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
