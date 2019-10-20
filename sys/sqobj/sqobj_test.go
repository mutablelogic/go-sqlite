package sqobj_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/sqlite"
	sq "github.com/djthorpe/sqlite"

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

type Device2 struct {
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

func (Device) TableName() string {
	return "device_table"
}

func (Device2) TableName() string {
	return "device_table"
}

/////////////////////////////////////////////////////////////////////////////

func Test_001(t *testing.T) {
	t.Log("Test_001")
}

func Test_002(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
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
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		if sqobj := app.ModuleInstance("db/sqobj").(sqlite.Objects); sqobj == nil {
			t.Fail()
		} else if sqconn := sqobj.Conn(); sqconn == nil {
			t.Fail()
		} else if sqlang := sqobj.Lang(); sqlang == nil {
			t.Fail()
		} else {
			t.Log(sqobj)
		}
	}
}

func Test_Register_004(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		db := app.ModuleInstance("db/sqobj").(sqlite.Objects)

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

func Test_Register_005(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		db := app.ModuleInstance("db/sqobj").(sqlite.Objects)
		if class, err := db.RegisterStruct(Device{}); err != nil {
			t.Error(err)
		} else if class.TableName() != "device_table" {
			t.Errorf("Unexpected class name: %v", class.Name())
		} else {
			t.Log(db)
		}
	}
}

func Test_Register_006(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		db := app.ModuleInstance("db/sqobj").(sqlite.Objects)

		// Empty or anonymous structs have unsupported type
		if _, err := db.RegisterStruct(struct{ A uint64 }{}); err == nil {
			t.Fail()
		} else if errors.Is(err, sqlite.ErrUnsupportedType) == false {
			t.Fail()
		} else {
			t.Log(err)
		}

		type B struct{}
		if _, err := db.RegisterStruct(B{}); err == nil {
			t.Fail()
		} else if errors.Is(err, sqlite.ErrUnsupportedType) == false {
			t.Fail()
		} else {
			t.Log(err)
		}

		type C struct{ uint64 }
		if _, err := db.RegisterStruct(C{}); err == nil {
			t.Fail()
		} else if errors.Is(err, sqlite.ErrUnsupportedType) == false {
			t.Fail()
		} else {
			t.Log(err)
		}

		type D struct{ E int64 }
		if _, err := db.RegisterStruct(D{}); err != nil {
			t.Fail()
		}
	}
}

func Test_Register_007(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		db := app.ModuleInstance("db/sqobj").(sqlite.Objects)
		if classA, err := db.RegisterStruct(&Device{}); err != nil {
			t.Error(err)
		} else if classB, err := db.RegisterStruct(&Device2{}); err != nil {
			t.Error(err)
		} else {
			t.Log(classA, classB)
		}
	}
}

func Test_Write_008(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
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

		obj := &Test{}

		if class, err := db.RegisterStruct(obj); err != nil {
			t.Error(err)
		} else if class.Name() != "Test" {
			t.Fail()
		} else if affected_rows, err := db.Write(sqlite.FLAG_INSERT, obj); err != nil {
			t.Error(err)
		} else if affected_rows != 1 {
			t.Error("Expected affected_rows to be 1")
		} else if obj.RowId == 0 {
			t.Error("Expected RowId != 0")
		} else {
			t.Log(obj)
		}
	}
}

func Test_Write_009(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
		t.Fail()
	} else {
		type Test struct {
			A int `sql:"A,primary"`
			sqlite.Object
		}
		defer app.Close()

		obj1 := &Test{A: 1}
		obj2 := &Test{A: 2}

		// Write two objects
		if class, err := db.RegisterStruct(&Test{}); err != nil {
			t.Error(err)
		} else if class.Name() != "Test" {
			t.Fail()
		} else if affected_rows, err := db.Write(sqlite.FLAG_INSERT, obj1, obj2); err != nil {
			t.Error(err)
		} else if affected_rows != 2 {
			t.Error("Expected affected_rows to be 2")
		} else if obj1.RowId == 0 || obj2.RowId == 0 || obj1.RowId == obj2.RowId {
			t.Error("Expected RowId != 0")
		} else {
			t.Log(obj1, obj2)
		}

		// Replace two objects
		obj1_id := obj1.RowId
		obj2_id := obj2.RowId

		if affected_rows, err := db.Write(sqlite.FLAG_INSERT|sqlite.FLAG_UPDATE, obj1, obj2); err != nil {
			t.Error(err)
		} else if affected_rows != 2 {
			t.Error("Expected affected_rows to be 2")
		} else if obj1.RowId == 0 || obj2.RowId == 0 || obj1.RowId == obj2.RowId {
			t.Error("Expected RowId != 0")
		} else if obj1.RowId != obj1_id || obj2.RowId != obj2_id {
			t.Error("Objects were inserted, not updated")
		} else {
			t.Log(obj1, obj2)
		}

		// Update two objects
		obj1.A = 100
		obj2.A = 200
		if affected_rows, err := db.Write(sqlite.FLAG_UPDATE, obj1, obj2); err != nil {
			t.Error(err)
		} else if affected_rows != 2 {
			t.Error("Expected affected_rows to be 2")
		} else if obj1.RowId == 0 || obj2.RowId == 0 || obj1.RowId == obj2.RowId {
			t.Error("Expected RowId != 0")
		} else if obj1.RowId != obj1_id || obj2.RowId != obj2_id {
			t.Error("Objects were inserted, not updated")
		} else {
			t.Log(obj1, obj2)
		}
	}
}

func Test_Write_010(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
		t.Fail()
	} else {
		type Test struct {
			A int `sql:"A,primary"`
			sqlite.Object
		}
		defer app.Close()

		obj := &Test{A: 1}

		// Update object which isn't in database
		if _, err := db.RegisterStruct(&Test{}); err != nil {
			t.Error(err)
		} else if _, err := db.Write(sqlite.FLAG_UPDATE, obj); errors.Is(err, gopi.ErrOutOfOrder) == false {
			t.Error("Expected Out of Order error")
		}
	}
}

func Test_Write_011(t *testing.T) {
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

func Test_Write_012(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
			t.Fail()
		} else {
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

func Test_Write_013(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		db := app.ModuleInstance("db/sqobj").(sqlite.Objects)

		type Test struct {
			sqlite.Object
			ID   int    `sql:"id,primary"`
			Name string `sql:"name"`
		}

		// Register device
		if _, err := db.RegisterStruct(Test{}); err != nil {
			t.Error(err)
		}

		test100 := &Test{ID: 100, Name: "test100"}

		// Write a row (insert)
		if affected_rows, err := db.Write(sqlite.FLAG_INSERT, &test100); err != nil {
			t.Error(err)
		} else if affected_rows != 1 {
			t.Fail()
		} else if test100.RowId != 100 {
			t.Fail()
		} else {
			t.Log(affected_rows)
		}

		// Write a row (insert) - should fail
		if _, err := db.Write(sqlite.FLAG_INSERT, &test100); err == nil {
			t.Error("Expected second insert to fail")
		}

		// Write a row (insert or update)
		if affected_rows, err := db.Write(sqlite.FLAG_INSERT|sqlite.FLAG_UPDATE, &test100); err != nil {
			t.Error(err)
		} else if affected_rows != 1 {
			t.Fail()
		} else if test100.RowId != 100 {
			t.Fail()
		} else {
			t.Log(affected_rows)
		}

		// Write a row (update)
		if affected_rows, err := db.Write(sqlite.FLAG_UPDATE, &test100); err != nil {
			t.Error(err)
		} else if affected_rows != 1 {
			t.Fail()
		} else if test100.RowId != 100 {
			t.Fail()
		} else {
			t.Log(affected_rows)
		}
	}
}

func Test_Delete_014(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		db := app.ModuleInstance("db/sqobj").(sqlite.Objects)

		type Test struct {
			sqlite.Object
			ID   int    `sql:"id,primary"`
			Name string `sql:"name"`
		}

		// Register
		if _, err := db.RegisterStruct(Test{}); err != nil {
			t.Error(err)
		}

		test100 := &Test{ID: 100, Name: "test100"}

		// Write a row (insert)
		if affected_rows, err := db.Write(sqlite.FLAG_INSERT, &test100); err != nil {
			t.Error(err)
		} else if affected_rows != 1 {
			t.Fail()
		} else if test100.RowId != 100 {
			t.Fail()
		} else {
			t.Log(affected_rows)
		}

		// Delete the row
		if affected_rows, err := db.Delete(&test100); err != nil {
			t.Error(err)
		} else if affected_rows != 1 {
			t.Fail()
		} else if test100.RowId != 0 {
			t.Fail()
		}

		// Delete the row again
		if _, err := db.Delete(&test100); errors.Is(err, gopi.ErrOutOfOrder) == false {
			t.Error("Expected out of order error")
		}
	}
}

func Test_Delete_015(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		db := app.ModuleInstance("db/sqobj").(sqlite.Objects)

		type Test struct {
			ID   int    `sql:"id,primary"`
			Name string `sql:"name"`
		}

		// Register
		if _, err := db.RegisterStruct(Test{}); err != nil {
			t.Error(err)
		}

		test100 := &Test{ID: 100, Name: "test100"}

		// Write a row (insert)
		if affected_rows, err := db.Write(sqlite.FLAG_INSERT, &test100); err != nil {
			t.Error(err)
		} else if affected_rows != 1 {
			t.Fail()
		} else {
			t.Log(affected_rows)
		}

		// Update the row (update)
		if affected_rows, err := db.Write(sqlite.FLAG_UPDATE, &test100); err != nil {
			t.Error(err)
		} else if affected_rows != 1 {
			t.Fail()
		} else {
			t.Log(affected_rows)
		}

		// Delete the row
		if affected_rows, err := db.Delete(&test100); err != nil {
			t.Error(err)
		} else if affected_rows != 1 {
			t.Fail()
		}
	}
}

func Test_Count_016(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		db := app.ModuleInstance("db/sqobj").(sqlite.Objects)
		class, err := db.RegisterStruct(&Device{})
		if err != nil {
			t.Error(err)
		} else if count, err := db.Count(class); err != nil {
			t.Error(err)
		} else if count != 0 {
			t.Error("Unexpected return from COUNT(*)")
		}

		// Insert 100 objects
		for i := 0; i < 100; i++ {
			if rows, err := db.Write(sq.FLAG_INSERT, &Device{}); err != nil {
				t.Error(err)
			} else if rows != 1 {
				t.Error("Unexpected affected rows")
			}
		}

		// Check count
		if count, err := db.Count(class); err != nil {
			t.Error(err)
		} else if count != 100 {
			t.Error("Unexpected return from COUNT(*), got", count, " (expected 100)")
		}
	}
}

/*
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

func Test_Reflect_009(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		if db := app.ModuleInstance("db/sqobj").(sqlite.Objects); db == nil {
			t.Fail()
		} else {
			type test_a struct {
				sqlite.Object
				A int64
			}
			type test_b struct {
				B int64
			}
			type test_c struct {
				C int64 `sql:"C,primary"`
			}
			type test_d struct {
				D int64  `sql:"D,primary"`
				E string `sql:"E,primary"`
			}
			type test_f struct {
				F int64  `sql:"F,primary"`
				G string `sql:"G,primary"`
				sqlite.Object
			}
			if class, err := db.RegisterStruct(&test_a{}); err != nil {
				t.Error(err)
			} else if keys := class.Keys(); len(keys) != 1 || keys[0] != "_rowid_" {
				t.Error("Keys() method returned unexpected result")
			}

			if class, err := db.RegisterStruct(&test_b{}); err != nil {
				t.Error(err)
			} else if keys := class.Keys(); len(keys) != 0 {
				t.Error("Keys() method returned unexpected result")
			}

			if class, err := db.RegisterStruct(&test_c{}); err != nil {
				t.Error(err)
			} else if keys := class.Keys(); len(keys) != 1 || keys[0] != "C" {
				t.Error("Keys() method returned unexpected result")
			}

			if class, err := db.RegisterStruct(&test_d{}); err != nil {
				t.Error(err)
			} else if keys := class.Keys(); len(keys) != 2 || keys[0] != "D" || keys[1] != "E" {
				t.Error("Keys() method returned unexpected result")
			}

			if class, err := db.RegisterStruct(&test_f{}); err != nil {
				t.Error(err)
			} else if keys := class.Keys(); len(keys) != 1 || keys[0] != "_rowid_" {
				t.Error("Keys() method returned unexpected result")
			}

		}
	}
}

func Test_Read_013(t *testing.T) {
	config := gopi.NewAppConfig("db/sqobj")
	if app, err := gopi.NewAppInstance(config); err != nil {
		t.Fatal(err)
	} else {
		defer app.Close()
		db := app.ModuleInstance("db/sqobj").(sqlite.Objects)

		type Test struct {
			sqlite.Object
			Name string `sql:"name"`
		}

		// Register and count number of objects in database
		cls, err := db.RegisterStruct(Test{})
		if err != nil {
			t.Error(err)
		} else if count, err := db.Count(cls); err != nil {
			t.Error(err)
		} else if count != 0 {
			t.Error("Expected count == 0")
		}

		// Add in two objects
		test := &Test{Name: "test"}

		// Write a row (insert)
		if affected_rows, err := db.Write(sqlite.FLAG_INSERT, test, test); err != nil {
			t.Error(err)
		} else if affected_rows != 2 {
			t.Fail()
		}

		// Count objects again
		if count, err := db.Count(cls); err != nil {
			t.Error(err)
		} else if count != 2 {
			t.Error("Expected count == 2")
		}
	}
}
*/
