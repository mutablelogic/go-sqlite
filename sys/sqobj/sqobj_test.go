package sqobj_test

import (
	"testing"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/sqlite"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/sqlite/sys/sqlite"
	_ "github.com/djthorpe/sqlite/sys/sqobj"
)

func Test_001(t *testing.T) {
	t.Log("Test_001")
}

func Test_002(t *testing.T) {
	config := gopi.NewAppConfig("db/sqlite", "db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else if driver_ := app.ModuleInstance("db/sqobj").(sqlite.Objects); driver_ == nil {
		t.Fail()
	} else {
		defer app.Close()
		t.Log(driver_)
	}
}

func Test_003(t *testing.T) {
	config := gopi.NewAppConfig("db/sqlite", "db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
		t.Fail()
	} else {
		defer app.Close()
		if class, err := db.RegisterStruct("test", struct{ A int }{}); err != nil {
			t.Error(err)
		} else if class.Name() != "test" {
			t.Errorf("Unexpected class name: %v", class.Name())
		} else {
			t.Log(class)
		}
	}
}

func Test_004(t *testing.T) {
	config := gopi.NewAppConfig("db/sqlite", "db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
		t.Fail()
	} else {
		type Test struct {
			A, B int
		}
		defer app.Close()
		if class, err := db.RegisterStruct("test", Test{}); err != nil {
			t.Error(err)
		} else if rowid, err := class.Insert(Test{}); err != nil {
			t.Error(err)
		} else if rowid == 0 {
			t.Error("Unexpected rowid", rowid)
		} else {
			t.Log("rowid=", rowid)
		}
	}
}

func Test_Reflect_001(t *testing.T) {
	config := gopi.NewAppConfig("db/sqlite", "db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
			t.Fail()
		} else if columns, err := db.ReflectStruct(struct{}{}); err != nil {
			t.Error(err)
		} else {
			t.Log(columns)
		}
	}
}

func Test_Reflect_002(t *testing.T) {
	config := gopi.NewAppConfig("db/sqlite", "db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
			t.Fail()
		} else {
			if columns, err := db.ReflectStruct(struct{ a int }{}); err != nil {
				t.Error(err)
			} else if len(columns) != 0 {
				t.Error("Expected zero returned columns")
			}

			if columns, err := db.ReflectStruct(struct{ A int }{}); err != nil {
				t.Error(err)
			} else if len(columns) != 1 {
				t.Error("Expected one returned columns")
			} else {
				t.Log(columns)
			}

			if columns, err := db.ReflectStruct(struct{ A, B int }{}); err != nil {
				t.Error(err)
			} else if len(columns) != 2 {
				t.Error("Expected two returned columns")
			} else {
				t.Log(columns)
			}

			if columns, err := db.ReflectStruct(struct {
				A int `sql:"test"`
			}{}); err != nil {
				t.Error(err)
			} else if len(columns) != 1 {
				t.Error("Expected two returned columns", columns)
			} else if columns[0].Name() != "test" {
				t.Error("Expected column name 'test'", columns)
			} else {
				t.Log(columns)
			}

			if columns, err := db.ReflectStruct(struct {
				A int `sql:",nullable"`
			}{}); err != nil {
				t.Error(err)
			} else if len(columns) != 1 {
				t.Error("Expected one returned columns", columns)
			} else if columns[0].Name() != "A" {
				t.Error("Expected column name 'A'", columns)
			} else if columns[0].Nullable() != true {
				t.Error("Expected column nullable", columns)
			} else {
				t.Log(columns)
			}

			if columns, err := db.ReflectStruct(struct {
				A string `sql:"TEST WITH SPACES,nullable,bool"`
			}{}); err != nil {
				t.Error(err)
			} else if len(columns) != 1 {
				t.Error("Expected one returned column", columns)
			} else if columns[0].Name() != "TEST WITH SPACES" {
				t.Error("Expected column name 'TEST WITH SPACES'", columns)
			} else if columns[0].Nullable() != true {
				t.Error("Expected column nullable", columns)
			} else if columns[0].DeclType() != "BOOL" {
				t.Error("Expected column type BOOL", columns)
			} else {
				t.Log(columns)
			}
		}
	}
}

func Test_Reflect_003(t *testing.T) {
	config := gopi.NewAppConfig("db/sqlite", "db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
			t.Fail()
		} else {
			if columns, err := db.ReflectStruct(struct {
				A int `sql:"test,primary"`
			}{}); err != nil {
				t.Error(err)
			} else if len(columns) != 1 {
				t.Error("Expected two returned columns", columns)
			} else if columns[0].Name() != "test" {
				t.Error("Expected column name 'test'", columns)
			} else if columns[0].PrimaryKey() != true {
				t.Error("Expected column 'test' with primary key", columns)
			} else {
				t.Log(columns)
			}
		}
	}
}

func Test_Reflect_004(t *testing.T) {
	config := gopi.NewAppConfig("db/sqlite", "db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
			t.Fail()
		} else if sqlite := app.ModuleInstance("db/sqlite").(sqlite.Connection); sqlite == nil {
			t.Fail()
		} else {
			if columns, err := db.ReflectStruct(struct {
				A int `sql:"a,primary"`
				B int `sql:"b,primary,nullable"`
			}{}); err != nil {
				t.Error(err)
			} else if len(columns) != 2 {
				t.Error("Expected two returned columns", columns)
			} else if columns[0].Name() != "a" {
				t.Error("Expected column name 'a'", columns)
			} else if columns[0].PrimaryKey() != true {
				t.Error("Expected column 'b' with primary key", columns)
			} else if columns[1].Name() != "b" {
				t.Error("Expected column name 'b'", columns)
			} else if columns[1].PrimaryKey() != true {
				t.Error("Expected column 'b' with primary key", columns)
			} else if create := sqlite.NewCreateTable("test", columns...); create == nil {
				t.Fail()
			} else if query := create.Query(sqlite); query == "" {
				t.Fail()
			} else if query != "CREATE TABLE test (a INTEGER NOT NULL,b INTEGER,PRIMARY KEY (a,b))" {
				t.Errorf("Unexpected query %v", query)
			} else {
				t.Log(query)
			}
		}
	}
}
