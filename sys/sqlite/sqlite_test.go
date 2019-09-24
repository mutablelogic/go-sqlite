package sqlite_test

import (
	"math"
	"testing"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
	sqlite "github.com/djthorpe/sqlite/sys/sqlite"
)

func Test_001(t *testing.T) {
	t.Log("Test_001")
}
func Test_002(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else if err := driver.Close(); err != nil {
		t.Error(err)
	}
}

func Test_003(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		defer driver.Close()
		if sqlite, ok := driver.(sq.Connection); !ok {
			t.Error("Cannot cast connection object")
			_ = driver.(sq.Connection)
		} else {
			t.Log(sqlite)
		}
	}
}

func Test_004(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if tables := driver_.TablesEx("", false); tables == nil {
			t.Error("Expected TablesEx to return empty slice, not nil")
		} else if len(tables) != 0 {
			t.Error("Expected Tables to return no strings")
		}
	}
}

func Test_005(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a,b)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if tables := driver_.TablesEx("", false); tables == nil {
			t.Error("Expected Tables to return empty slice")
		} else if len(tables) != 1 {
			t.Error("Expected Tables to return a single value")
		} else if tables[0] != "test" {
			t.Errorf("Unexpected return value from Tables: %v", tables)
		}
	}
}

func Test_006(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a,b); CREATE TABLE test2 (c,d)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if tables := driver_.TablesEx("", false); tables == nil {
			t.Error("Expected Tables to return empty slice")
		} else if len(tables) != 2 {
			t.Error("Expected Tables to return a single value")
		} else if tables[0] != "test" {
			t.Errorf("Unexpected return value from Tables: %v", tables)
		} else if tables[1] != "test2" {
			t.Errorf("Unexpected return value from Tables: %v", tables)
		}
	}
}

func Test_007(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a,b)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if results, err := driver_.DoOnce("INSERT INTO test VALUES (1,2)"); err != nil {
			t.Error(err)
		} else if results.RowsAffected != 1 {
			t.Errorf("Unexpected RowsAffected value: %v", results)
		}
	}
}

func Test_008(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a,b)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if results, err := driver_.DoOnce("INSERT INTO test VALUES (1,2),(3,4)"); err != nil {
			t.Error(err)
		} else if results.RowsAffected != 2 {
			t.Errorf("Unexpected RowsAffected value: %v", results)
		}
	}
}

func Test_009(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a INTEGER,b INTEGER)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if _, err := driver_.DoOnce("INSERT INTO test VALUES (1,2),(3,4)"); err != nil {
			t.Error(err)
		} else if st := driver_.NewStatement("SELECT a,b FROM test"); st == nil {
			t.Error("<nil> Statement")
		} else if rs, err := driver_.Query(st); err != nil {
			t.Error(err)
		} else {
			t.Log(rs)
			for {
				row := rs.Next()
				if row == nil {
					break
				}
				t.Log(row)
			}
		}
	}
}

func Test_010(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a timestamp)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if _, err := driver_.DoOnce("INSERT INTO test VALUES (?)", time.Now()); err != nil {
			t.Error(err)
		} else if st := driver_.NewStatement("SELECT a FROM test"); st == nil {
			t.Error("<nil> Statement")
		} else if rs, err := driver_.Query(st); err != nil {
			t.Error(err)
		} else {
			for {
				row := rs.Next()
				if row == nil {
					break
				}
				t.Log(row[0].String())
				t.Log(row[0].Timestamp())
			}
		}
	}
}

func Test_011(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a integer,b integer)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if insert := driver_.NewStatement("INSERT INTO test (a,b) VALUES (?,?)"); insert == nil {
			t.Error("<nil> Statement")
		} else if _, err := driver_.DoOnce("INSERT INTO test (a,b) VALUES (0,12)"); err != nil {
			t.Error(err)
		} else if _, err := driver_.Do(insert, 1, 34); err != nil {
			t.Error(err)
		} else if _, err := driver_.DoOnce("INSERT INTO test (a,b) VALUES (2,NULL)"); err != nil {
			t.Error(err)
		} else if _, err := driver_.DoOnce("INSERT INTO test (a,b) VALUES (3,?)", nil); err != nil {
			t.Error(err)
		} else if rs, err := driver_.QueryOnce("SELECT a,b FROM test ORDER BY a"); err != nil {
			t.Error(err)
		} else {
			for i := 0; true; i++ {
				row := rs.Next()
				if row == nil {
					break
				}
				t.Log(sq.RowString(row))
				switch i {
				case 0:
					if len(row) != 2 || row[0].Int() != 0 || row[1].Int() != 12 {
						t.Errorf("Unexpected row[%v]: %v", i, sq.RowString(row))
					}
				case 1:
					if len(row) != 2 || row[0].Int() != 1 || row[1].Int() != 34 {
						t.Errorf("Unexpected row[%v]: %v", i, sq.RowString(row))
					}
				case 2:
					if len(row) != 2 || row[0].Int() != 2 || row[1].IsNull() == false {
						t.Errorf("Unexpected row[%v]: %v", i, sq.RowString(row))
					}
				case 3:
					if len(row) != 2 || row[0].Int() != 3 || row[1].IsNull() == false {
						t.Errorf("Unexpected row[%v]: %v", i, sq.RowString(row))
					}
				}
			}
		}
	}
}

func Test_012(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a blob)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if _, err := driver_.DoOnce("INSERT INTO test VALUES (?)", "hello"); err != nil {
			t.Error(err)
		} else if _, err := driver_.DoOnce("INSERT INTO test VALUES (?)", []byte{0, 1, 2, 3, 4}); err != nil {
			t.Error(err)
		} else if rs, err := driver_.QueryOnce("SELECT a FROM test ORDER BY a"); err != nil {
			t.Error(err)
		} else {
			t.Log(rs)
			for {
				row := rs.Next()
				if row == nil {
					break
				}
				t.Log(sq.RowString(row))
			}
		}
	}
}

func Test_013(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if schemas := driver_.Schemas(); schemas == nil {
			t.Fail()
		} else {
			t.Log(schemas)
		}
	}
}

func Test_014(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TEMPORARY TABLE test (a integer)"); err != nil {
			t.Error(err)
		} else if schemas := driver_.Schemas(); schemas == nil {
			t.Fail()
		} else if schemas[0] != "main" {
			t.Fail()
		} else if schemas[1] != "temp" {
			t.Fail()
		} else if tables := driver_.TablesEx("", false); tables == nil {
			t.Fail()
		} else if len(tables) != 0 {
			t.Fail()
		} else if tables := driver_.TablesEx("temp", false); tables == nil {
			t.Fail()
		} else if len(tables) != 1 || tables[0] != "test" {
			t.Fail()
		} else if tables := driver_.TablesEx("", true); tables == nil {
			t.Fail()
		} else if len(tables) != 1 || tables[0] != "test" {
			t.Fail()
		}
	}
}

func Test_Attach_015(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		if err := driver_.Attach("main", ":memory:"); err == nil {
			t.Fail()
		} else {
			t.Log(err)
		}

		if err := driver_.Attach("test", ":memory:"); err != nil {
			t.Error(err)
		} else if schemas := driver_.Schemas(); len(schemas) != 2 {
			t.Fail()
		} else if schemas[0] != "main" && schemas[1] != "test" {
			t.Fail()
		}

		if err := driver_.Attach("test", ":memory:"); err == nil {
			t.Fail()
		} else {
			t.Log(err)
		}

		if err := driver_.Attach("temp", ":memory:"); err == nil {
			t.Fail()
		} else {
			t.Log(err)
		}

	}
}

func Test_Attach_016(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		if err := driver_.Attach("test", ":memory:"); err != nil {
			t.Error(err)
		}
		if err := driver_.Detach("test"); err != nil {
			t.Error(err)
		}
		if err := driver_.Detach("main"); err == nil {
			t.Fail()
		}
	}
}

func Test_Txn_017(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		if st := driver_.NewStatement("test"); st == nil {
			t.Fail()
		} else if st.Query() != "test" {
			t.Fail()
		}
	}
}

func Test_Txn_018(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		if st := driver_.NewStatement("SELECT 1"); st == nil {
			t.Fail()
		} else if rs, err := driver_.Query(st); err != nil {
			t.Error(err)
		} else if rs == nil {
			t.Fail()
		} else if cols := rs.Columns(); len(cols) != 1 {
			t.Fail()
		} else {
			t.Log(cols)
		}
	}
}

func Test_Txn_019(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		if err := driver_.Txn(func(tx sq.Transaction) error {
			_, err := tx.DoOnce("SELECT 1")
			return err
		}); err != nil {
			t.Error(err)
		}
	}
}

func Test_Txn_020(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		// Attempt to make a transaction within a transaction, should fail
		// with out-of-order error
		if err := driver_.Txn(func(tx sq.Transaction) error {
			return driver_.Txn(func(tx sq.Transaction) error {
				return nil
			})
		}); err != gopi.ErrOutOfOrder {
			t.Fail()
		}
	}
}

func Test_Txn_021(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		// Create a table
		if _, err := driver_.DoOnce("CREATE TABLE test (A int)"); err != nil {
			t.Fatal(err)
		}

		// Insert two records with a transaction
		if err := driver_.Txn(func(tx sq.Transaction) error {
			if _, err := driver_.DoOnce("INSERT INTO test DEFAULT VALUES"); err != nil {
				return err
			}
			if _, err := driver_.DoOnce("INSERT INTO test DEFAULT VALUES"); err != nil {
				return err
			}
			return nil
		}); err != nil {
			t.Error(err)
		}
		// Check number of inserted records
		if rs, err := driver_.QueryOnce("SELECT COUNT(*) FROM test"); err != nil {
			t.Fatal(err)
		} else if row := rs.Next(); len(row) != 1 || row[0].String() != "2" {
			t.Fail()
		}
		// Delete records
		if err := driver_.Txn(func(tx sq.Transaction) error {
			if _, err := driver_.DoOnce("DELETE FROM test"); err != nil {
				return err
			}
			return gopi.ErrAppError
		}); err != gopi.ErrAppError {
			t.Fail()
		}
		// Check number of inserted records
		if rs, err := driver_.QueryOnce("SELECT COUNT(*) FROM test"); err != nil {
			t.Fatal(err)
		} else if row := rs.Next(); len(row) != 1 || row[0].String() != "2" {
			t.Fail()
		}
	}
}

func Test_Types_022(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		// Create a table
		if _, err := driver_.DoOnce("CREATE TABLE test (A INTEGER NOT NULL)"); err != nil {
			t.Fatal(err)
		}

		// Insert and check int64 values
		values := []int64{
			0,
			-100,
			100,
			int64(math.MinInt64),
			int64(math.MaxInt64),
		}

		for _, value := range values {
			if r, err := driver_.DoOnce("INSERT INTO test VALUES (?)", value); err != nil {
				t.Error(err)
			} else if rs, err := driver_.QueryOnce("SELECT * FROM test WHERE _rowid_=?", r.LastInsertId); err != nil {
				t.Error(err)
			} else if row := rs.Next(); row == nil {
				t.Fail()
			} else if row[0].IsNull() {
				t.Fail()
			} else if row[0].Int() != value {
				t.Error("Unexpected value", row, "(expected", value, ")")
			} else {
				t.Log(row)
			}
		}
	}
}

func Test_Types_023(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		// Create a table
		if _, err := driver_.DoOnce("CREATE TABLE test (A BOOL NOT NULL)"); err != nil {
			t.Fatal(err)
		}

		// Insert and check int64 values
		values := []bool{
			false,
			true,
		}

		for _, value := range values {
			if r, err := driver_.DoOnce("INSERT INTO test VALUES (?)", value); err != nil {
				t.Error(err)
			} else if rs, err := driver_.QueryOnce("SELECT * FROM test WHERE _rowid_=?", r.LastInsertId); err != nil {
				t.Error(err)
			} else if row := rs.Next(); row == nil {
				t.Fail()
			} else if row[0].IsNull() {
				t.Fail()
			} else if row[0].Bool() != value {
				t.Error("Unexpected value", row, "(expected", value, ")")
			} else {
				t.Log(row)
			}
		}
	}
}

func Test_Types_024(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		// Create a table
		if _, err := driver_.DoOnce("CREATE TABLE test (A FLOAT NOT NULL)"); err != nil {
			t.Fatal(err)
		}

		// Insert and check int64 values
		values := []float64{
			0,
			1.0,
			-1.0,
			math.Pi,
			-math.Pi,
			math.E,
			-math.E,
		}

		for _, value := range values {
			if r, err := driver_.DoOnce("INSERT INTO test VALUES (?)", value); err != nil {
				t.Error(err)
			} else if rs, err := driver_.QueryOnce("SELECT * FROM test WHERE _rowid_=?", r.LastInsertId); err != nil {
				t.Error(err)
			} else if row := rs.Next(); row == nil {
				t.Fail()
			} else if row[0].IsNull() {
				t.Fail()
			} else if row[0].Float() != value {
				t.Error("Unexpected value", row, "(expected", value, ")")
			} else {
				t.Log(row)
			}
		}
	}
}

func Test_Types_026(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{
		Location: "UTC",
	}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		// Create a table
		if _, err := driver_.DoOnce("CREATE TABLE test (A TIMESTAMP NOT NULL)"); err != nil {
			t.Fatal(err)
		}

		// Insert and check int64 values
		values := []time.Time{
			time.Time{},
			time.Now(),
			time.Now().Add(100 * time.Hour),
			time.Now().Add(-365 * time.Hour),
		}

		for _, value := range values {
			if r, err := driver_.DoOnce("INSERT INTO test VALUES (?)", value); err != nil {
				t.Error(err)
			} else if rs, err := driver_.QueryOnce("SELECT * FROM test WHERE _rowid_=?", r.LastInsertId); err != nil {
				t.Error(err)
			} else if row := rs.Next(); row == nil {
				t.Fail()
			} else if row[0].IsNull() {
				t.Fail()
			} else if row[0].Timestamp() != value {
				t.Error("Unexpected value", row, "(expected", value, ")")
			} else {
				t.Log(row)
			}
		}
	}
}

func Test_Types_027(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Database{
		Location: "Local",
	}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()

		// Create a table
		if _, err := driver_.DoOnce("CREATE TABLE test (A TIMESTAMP NOT NULL)"); err != nil {
			t.Fatal(err)
		}

		// Insert and check int64 values
		values := []time.Time{
			time.Time{},
			time.Now(),
			time.Now().Add(100 * time.Hour),
			time.Now().Add(-365 * time.Hour),
		}

		for _, value := range values {
			if r, err := driver_.DoOnce("INSERT INTO test VALUES (?)", value); err != nil {
				t.Error(err)
			} else if rs, err := driver_.QueryOnce("SELECT * FROM test WHERE _rowid_=?", r.LastInsertId); err != nil {
				t.Error(err)
			} else if row := rs.Next(); row == nil {
				t.Fail()
			} else if row[0].IsNull() {
				t.Fail()
			} else if row[0].Timestamp() != value {
				t.Error("Unexpected value", row, "(expected", value, ")")
			} else {
				t.Log(row)
			}
		}
	}
}
