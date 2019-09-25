package sqobj_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/sqlite"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/sqlite/sys/sqlite"
	_ "github.com/djthorpe/sqlite/sys/sqobj"
)

/////////////////////////////////////////////////////////////////////////////

type Device struct {
	sqlite.Object

	ID          int       `sql:"device_id"`
	Name        string    `sql:"name"`
	DateAdded   time.Time `sql:"date_added"`
	DateUpdated time.Time `sql:"date_updated,nullable"`
	Enabled     bool      `sql:"enabled"`
}

func (this *Device) String() string {
	return fmt.Sprintf("<Device>{ ID=%v Name=%v Object=%v }", this.ID, strconv.Quote(this.Name), this.Object.String())
}

/////////////////////////////////////////////////////////////////////////////

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
		type Test struct {
			A int
		}
		if class, err := db.RegisterStruct(Test{}); err != nil {
			t.Error(err)
		} else if class.Name() != "Test" {
			t.Errorf("Unexpected class name: %v", class.Name())
		} else {
			t.Log(db)
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
			sqlite.Object
		}
		defer app.Close()
		if class, err := db.RegisterStruct(Test{}); err != nil {
			t.Error(err)
		} else if class.Name() != "Test" {
			t.Fail()
		} else if affected_rows, err := db.Write(sqlite.FLAG_INSERT, &Test{}); err != nil {
			t.Error(err)
		} else if affected_rows != 1 {
			t.Fail()
		}
	}
}

func Test_Reflect_005(t *testing.T) {
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

func Test_Reflect_006(t *testing.T) {
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

func Test_Reflect_007(t *testing.T) {
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

func Test_Reflect_008(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
			t.Fail()
		} else if lang_ := app.ModuleInstance("db/sqlang").(sqlite.Language); lang_ == nil {
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
			} else if create := lang_.NewCreateTable("test", columns...); create == nil {
				t.Fail()
			} else if query := create.Query(); query == "" {
				t.Fail()
			} else if query != "CREATE TABLE test (a INTEGER NOT NULL,b INTEGER,PRIMARY KEY (a,b))" {
				t.Errorf("Unexpected query %v", query)
			} else {
				t.Log(query)
			}
		}
	}
}

func Test_Insert_009(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
			t.Fail()
		} else {

			if class, err := db.RegisterStruct(Device{}); err != nil {
				t.Error(err)
			} else {
				t.Log(class)

				// In this case, the primary key is auto-generated and the first two rows
				// will have rowid of 1 and 2
				device100 := &Device{ID: 100, Name: "Device100"}
				device101 := &Device{ID: 101, Name: "Device101"}

				if affected_rows, err := db.Write(sqlite.FLAG_INSERT, device100); err != nil {
					t.Error(err)
				} else if affected_rows != 1 {
					t.Error()
				} else if device100.RowId != 1 {
					t.Error()
				} else {
					t.Log(device100, affected_rows)
				}

				if affected_rows, err := db.Write(sqlite.FLAG_INSERT, device101); err != nil {
					t.Error(err)
				} else if affected_rows != 1 {
					t.Error("affected_rows != 1")
				} else if device101.RowId != 2 {
					t.Error("device101.RowId != 2", device101)
				} else {
					t.Log(device101, affected_rows)
				}

			}
		}
	}
}

func Test_Insert_010(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
			t.Fail()
		} else {

			type Device struct {
				ID          int       `sql:"device_id,primary"`
				Name        string    `sql:"name"`
				DateAdded   time.Time `sql:"date_added"`
				DateUpdated time.Time `sql:"date_updated,nullable"`
				Enabled     bool      `sql:"enabled"`
			}

			if _, err := db.RegisterStruct(Device{}); err != nil {
				t.Error(err)
			} else {
				// In this case, the primary key is row ID auto-generated and the first two rows
				// will have rowid of 100 and 101
				if affected_rows, err := db.Write(sqlite.FLAG_INSERT, &Device{ID: 100}); err != nil {
					t.Error(err)
				} else {
					t.Log(affected_rows)
				}

				if affected_rows, err := db.Write(sqlite.FLAG_INSERT, &Device{ID: 101}); err != nil {
					t.Error(err)
				} else {
					t.Log(affected_rows)
				}

			}
		}
	}
}
