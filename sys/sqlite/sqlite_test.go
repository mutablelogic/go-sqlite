package sqlite_test

import (
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
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else if err := driver.Close(); err != nil {
		t.Error(err)
	}
}

func Test_003(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
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
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if tables := driver_.Tables(); tables == nil {
			t.Error("Expected Tables to return empty slice")
		}
	}
}

func Test_005(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a,b)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if tables := driver_.Tables(); tables == nil {
			t.Error("Expected Tables to return empty slice")
		} else if len(tables) != 1 {
			t.Error("Expected Tables to return a single value")
		} else if tables[0] != "test" {
			t.Errorf("Unexpected return value from Tables: %v", tables)
		}
	}
}

func Test_006(t *testing.T) {
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a,b); CREATE TABLE test2 (c,d)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if tables := driver_.Tables(); tables == nil {
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
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
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
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
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
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a INTEGER,b INTEGER)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if _, err := driver_.DoOnce("INSERT INTO test VALUES (1,2),(3,4)"); err != nil {
			t.Error(err)
		} else if st, err := driver_.Prepare("SELECT a,b FROM test"); err != nil {
			t.Error(err)
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
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a timestamp)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if _, err := driver_.DoOnce("INSERT INTO test VALUES (?)", time.Now()); err != nil {
			t.Error(err)
		} else if st, err := driver_.Prepare("SELECT a FROM test"); err != nil {
			t.Error(err)
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
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
		t.Error(err)
	} else {
		driver_ := driver.(sq.Connection)
		defer driver_.Close()
		if _, err := driver_.DoOnce("CREATE TABLE test (a integer,b integer)"); err != nil {
			t.Error("Expected Tables to return empty slice")
		} else if insert, err := driver_.Prepare("INSERT INTO test (a,b) VALUES (?,?)"); err != nil {
			t.Error(err)
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
	if driver, err := gopi.Open(sqlite.Config{}, nil); err != nil {
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
