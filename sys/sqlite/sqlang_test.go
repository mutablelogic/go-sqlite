package sqlite_test

import (
	"strconv"
	"testing"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
	sqlite "github.com/djthorpe/sqlite/sys/sqlite"
)

func Test_Lang_001(t *testing.T) {
	t.Log("Test_Lang_001")
}

func Test_Create_002(t *testing.T) {
	if lang, err := gopi.Open(sqlite.Language{}, nil); err != nil {
		t.Error(err)
	} else if lang_ := lang.(sq.Language); lang_ == nil {
		t.Fail()
	} else {
		defer lang_.Close()

		tests := []struct {
			f     func() sq.CreateTable
			query string
		}{
			{func() sq.CreateTable { return lang_.NewCreateTable("test") }, "CREATE TABLE test ()"},
			{func() sq.CreateTable { return lang_.NewCreateTable("test").Schema("test") }, "CREATE TABLE test.test ()"},
			{func() sq.CreateTable { return lang_.NewCreateTable("test").Temporary() }, "CREATE TEMPORARY TABLE test ()"},
			{func() sq.CreateTable { return lang_.NewCreateTable("test").IfNotExists() }, "CREATE TABLE IF NOT EXISTS test ()"},
			{func() sq.CreateTable { return lang_.NewCreateTable("test").WithoutRowID() }, "CREATE TABLE test () WITHOUT ROWID"},
		}

		for i, test := range tests {
			if statement := test.f(); statement == nil {
				t.Errorf("Test %v: nil value returned", i)
			} else if statement.Query() != test.query {
				t.Errorf("Test %v: Expected %v, got %v", i, strconv.Quote(test.query), strconv.Quote(statement.Query()))
			}
		}
	}
}

func Test_Create_003(t *testing.T) {
	if lang, err := gopi.Open(sqlite.Language{}, nil); err != nil {
		t.Error(err)
	} else if lang_ := lang.(sq.Language); lang_ == nil {
		t.Fail()
	} else if db, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else if db_ := db.(sq.Connection); lang_ == nil {
		t.Fail()
	} else {
		defer lang_.Close()
		defer db_.Close()

		if statement := lang_.NewCreateTable("test", db_.NewColumn("a", "TEXT", false, false), db_.NewColumn("b", "TEXT", true, false)); statement == nil {
			t.Error("Statement returned is nil")
		} else if statement.Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT)" {
			t.Errorf("Unexpected value, %v", statement.Query())
		}
	}
}

func Test_Create_004(t *testing.T) {
	if lang, err := gopi.Open(sqlite.Language{}, nil); err != nil {
		t.Error(err)
	} else if lang_ := lang.(sq.Language); lang_ == nil {
		t.Fail()
	} else if db, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else if db_ := db.(sq.Connection); lang_ == nil {
		t.Fail()
	} else {
		defer lang_.Close()
		defer db_.Close()

		if statement := lang_.NewCreateTable("test", db_.NewColumn("a", "TEXT", false, true), db_.NewColumn("b", "TEXT", true, false)); statement == nil {
			t.Error("Statement returned is nil")
		} else if statement.Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT,PRIMARY KEY (a))" {
			t.Errorf("Unexpected value, %v", statement.Query())
		}
	}
}

func Test_Create_005(t *testing.T) {
	if lang, err := gopi.Open(sqlite.Language{}, nil); err != nil {
		t.Error(err)
	} else if lang_ := lang.(sq.Language); lang_ == nil {
		t.Fail()
	} else if db, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else if db_ := db.(sq.Connection); lang_ == nil {
		t.Fail()
	} else {
		defer lang_.Close()
		defer db_.Close()

		if statement := lang_.NewCreateTable("test", db_.NewColumn("a", "TEXT", false, true), db_.NewColumn("b", "TEXT", true, false)); statement == nil {
			t.Error("Statement returned is nil")
		} else if statement.Unique("a", "b").Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT,PRIMARY KEY (a),UNIQUE (a,b))" {
			t.Errorf("Unexpected value, %v", statement.Query())
		}

		if statement := lang_.NewCreateTable("test", db_.NewColumn("a", "TEXT", false, false), db_.NewColumn("b", "TEXT", true, false)); statement == nil {
			t.Error("Statement returned is nil")
		} else if statement.Unique("a").Unique("b").Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT,UNIQUE (a),UNIQUE (b))" {
			t.Errorf("Unexpected value, %v", statement.Query())
		}

		if statement := lang_.NewCreateTable("test", db_.NewColumn("a", "TEXT", false, true), db_.NewColumn("b", "TEXT", true, true)); statement == nil {
			t.Error("Statement returned is nil")
		} else if statement.Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT,PRIMARY KEY (a,b))" {
			t.Errorf("Unexpected value, %v", statement.Query())
		}

		if statement := lang_.NewCreateTable("test", db_.NewColumn("a", "TEXT", false, true), db_.NewColumn("b", "TEXT", true, true)); statement == nil {
			t.Error("Statement returned is nil")
		} else if statement.Unique("a", "b").Query() != "CREATE TABLE test (a TEXT NOT NULL,b TEXT,PRIMARY KEY (a,b),UNIQUE (a,b))" {
			t.Errorf("Unexpected value, %v", statement.Query())
		}
	}
}

func Test_Create_006(t *testing.T) {
	if lang, err := gopi.Open(sqlite.Language{}, nil); err != nil {
		t.Error(err)
	} else if lang_ := lang.(sq.Language); lang_ == nil {
		t.Fail()
	} else if db, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else if db_ := db.(sq.Connection); lang_ == nil {
		t.Fail()
	} else {
		defer lang_.Close()
		defer db_.Close()

		if statement := lang_.NewCreateTable("test", db_.NewColumn("a", "TEXT", false, false), db_.NewColumn("b", "TEXT", true, false)); statement == nil {
			t.Error("Statement returned is nil")
		} else if _, err := db_.DoOnce(statement.Query()); err != nil {
			t.Error(err)
		}

		if _, err := db_.DoOnce(lang_.DropTable("test").Query()); err != nil {
			t.Error(err)
		}

		if statement := lang_.NewCreateTable("test", db_.NewColumn("a", "TEXT", false, false), db_.NewColumn("b", "TEXT", true, true)); statement == nil {
			t.Error("Statement returned is nil")
		} else if _, err := db_.DoOnce(statement.Query()); err != nil {
			t.Error(err)
		}

	}
}

func Test_Drop_007(t *testing.T) {
	if lang, err := gopi.Open(sqlite.Language{}, nil); err != nil {
		t.Error(err)
	} else if lang_ := lang.(sq.Language); lang_ == nil {
		t.Fail()
	} else {
		defer lang_.Close()

		if statement := lang_.DropTable("test"); statement.Query() != "DROP TABLE test" {
			t.Error("Unexpected query:", statement.Query())
		} else if statement.IfExists(); statement.Query() != "DROP TABLE IF EXISTS test" {
			t.Error("Unexpected query:", statement.Query())
		} else if statement.Schema("test"); statement.Query() != "DROP TABLE IF EXISTS test.test" {
			t.Error("Unexpected query:", statement.Query())
		}
	}
}

func Test_Insert_008(t *testing.T) {
	if lang, err := gopi.Open(sqlite.Language{}, nil); err != nil {
		t.Error(err)
	} else if lang_ := lang.(sq.Language); lang_ == nil {
		t.Fail()
	} else if db, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else if db_ := db.(sq.Connection); lang_ == nil {
		t.Fail()
	} else {
		defer lang_.Close()
		defer db_.Close()

		if column := db_.NewColumn("a", "TEST", false, false); column == nil {
			t.Fail()
		} else if create := lang_.NewCreateTable("test", column); create == nil {
			t.Fail()
		} else if _, err := db_.Do(create); err != nil {
			t.Error(err)
		} else if statement := lang_.Insert("test"); statement.Query() != "INSERT INTO test DEFAULT VALUES" {
			t.Error("Unexpected query:", statement.Query())
		} else if statement := lang_.Insert("test", "a"); statement.Query() != "INSERT INTO test (a) VALUES (?)" {
			t.Error("Unexpected query:", statement.Query())
		} else if statement := lang_.Insert("test").DefaultValues(); statement.Query() != "INSERT INTO test DEFAULT VALUES" {
			t.Error("Unexpected query:", statement.Query())
		} else if statement := lang_.Insert("test", "a", "b"); statement.Query() != "INSERT INTO test (a,b) VALUES (?,?)" {
			t.Error("Unexpected query:", statement.Query())
		}
	}
}

func Test_Query_009(t *testing.T) {
	if lang, err := gopi.Open(sqlite.Language{}, nil); err != nil {
		t.Error(err)
	} else if lang_ := lang.(sq.Language); lang_ == nil {
		t.Fail()
	} else {
		defer lang_.Close()

		if statement := lang_.NewSelect(nil); statement.Query() != "SELECT *" {
			t.Error("Unexpected query:", statement.Query())
		} else {
			t.Log(statement.Query())
		}

		if statement := lang_.NewSelect(nil).Distinct(); statement.Query() != "SELECT DISTINCT *" {
			t.Error("Unexpected query:", statement.Query())
		} else {
			t.Log(statement.Query())
		}

		if statement := lang_.NewSelect(nil).LimitOffset(0, 0); statement.Query() != "SELECT *" {
			t.Error("Unexpected query:", statement.Query())
		} else {
			t.Log(statement.Query())
		}

		if statement := lang_.NewSelect(nil).LimitOffset(0, 1); statement.Query() != "SELECT * OFFSET 1" {
			t.Error("Unexpected query:", statement.Query())
		} else {
			t.Log(statement.Query())
		}

		if statement := lang_.NewSelect(nil).LimitOffset(100, 0); statement.Query() != "SELECT * LIMIT 100" {
			t.Error("Unexpected query:", statement.Query())
		} else {
			t.Log(statement.Query())
		}

		if statement := lang_.NewSelect(nil).LimitOffset(100, 1); statement.Query() != "SELECT * LIMIT 100,1" {
			t.Error("Unexpected query:", statement.Query())
		} else {
			t.Log(statement.Query())
		}
	}
}

func Test_Query_010(t *testing.T) {
	if lang, err := gopi.Open(sqlite.Language{}, nil); err != nil {
		t.Error(err)
	} else if lang_ := lang.(sq.Language); lang_ == nil {
		t.Fail()
	} else {
		defer lang_.Close()

		if source := lang_.NewSource("column_a"); source == nil {
			t.Error("Unexpected <nil> returned from NewSource")
		} else if statement := lang_.NewSelect(source); statement.Query() != "SELECT * FROM column_a" {
			t.Error("Unexpected query:", statement.Query())
		} else {
			t.Log(statement.Query())
		}

		if source := lang_.NewSource("column_a").Schema("test"); source == nil {
			t.Error("Unexpected <nil> returned from NewSource")
		} else if statement := lang_.NewSelect(source); statement.Query() != "SELECT * FROM test.column_a" {
			t.Error("Unexpected query:", statement.Query())
		} else {
			t.Log(statement.Query())
		}

		if source := lang_.NewSource("column_a").Alias("a"); source == nil {
			t.Error("Unexpected <nil> returned from NewSource")
		} else if statement := lang_.NewSelect(source); statement.Query() != "SELECT * FROM column_a AS a" {
			t.Error("Unexpected query:", statement.Query())
		} else {
			t.Log(statement.Query())
		}

		if source := lang_.NewSource("column_a").Alias("a").Schema("test"); source == nil {
			t.Error("Unexpected <nil> returned from NewSource")
		} else if statement := lang_.NewSelect(source); statement.Query() != "SELECT * FROM test.column_a AS a" {
			t.Error("Unexpected query:", statement.Query())
		} else {
			t.Log(statement.Query())
		}

	}
}
