package sqlite_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	"time"

	// Modules
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"

	// Import into namespace
	. "github.com/djthorpe/go-sqlite/pkg/lang"
)

func Test_Resultset_001(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(N("foo").CreateTable(N("a").WithType("INTEGER"))); err != nil {
		t.Error(err)
	}

	// Prepare insert statement
	insert, err := db.Prepare(N("foo").Insert("a"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log(insert)
	}

	// Insert 100 int64 values
	n := 100
	v := make([]int64, n)
	for i := range v {
		v[i] = rand.Int63()
		if _, err := db.Exec(insert, v[i]); err != nil {
			t.Error(err)
		}
	}
	// Read rows - returns int64 values
	rows, err := db.Query(S(N("foo")))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= n; i++ {
		row := rows.Next()
		if i == n && row == nil {
			break
		} else if len(row) != 1 {
			t.Error("Unexpected row length", len(row), "in position", i)
		} else if d, ok := row[0].(int64); !ok {
			t.Errorf("Unexpected type %T in position %d", row[0], i)
		} else if d != v[i] {
			t.Error("Unexpected value", row[0], "expected", v[i], "in position", i)
		}
	}
}

func Test_Resultset_002(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(N("foo").CreateTable(N("a").WithType("TEXT"))); err != nil {
		t.Error(err)
	}

	// Prepare insert statement
	insert, err := db.Prepare(N("foo").Insert("a"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log(insert)
	}

	// Insert 100 string values
	n := 100
	v := make([]string, n)
	for i := range v {
		v[i] = fmt.Sprint(rand.Int63())
		if _, err := db.Exec(insert, v[i]); err != nil {
			t.Error(err)
		}
	}
	// Read rows - returns string values
	rows, err := db.Query(S(N("foo")))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= n; i++ {
		row := rows.Next()
		if i == n && row == nil {
			break
		} else if len(row) != 1 {
			t.Error("Unexpected row length", len(row), "in position", i)
		} else if d, ok := row[0].(string); !ok {
			t.Errorf("Unexpected type %T in position %d", row[0], i)
		} else if d != v[i] {
			t.Error("Unexpected value", row[0], "expected", v[i], "in position", i)
		}
	}
}

func Test_Resultset_003(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(N("foo").CreateTable(N("a").WithType("FLOAT"))); err != nil {
		t.Error(err)
	}

	// Prepare insert statement
	insert, err := db.Prepare(N("foo").Insert("a"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log(insert)
	}

	// Insert 100 string values
	n := 100
	v := make([]float64, n)
	for i := range v {
		v[i] = rand.Float64()
		if _, err := db.Exec(insert, v[i]); err != nil {
			t.Error(err)
		}
	}
	// Read rows - returns float64 values
	rows, err := db.Query(S(N("foo")))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= n; i++ {
		row := rows.Next()
		if i == n && row == nil {
			break
		} else if len(row) != 1 {
			t.Error("Unexpected row length", len(row), "in position", i)
		} else if d, ok := row[0].(float64); !ok {
			t.Errorf("Unexpected type %T in position %d", row[0], i)
		} else if d != v[i] {
			t.Error("Unexpected value", row[0], "expected", v[i], "in position", i)
		}
	}
}

func Test_Resultset_004(t *testing.T) {
	db, err := sqlite.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(N("foo").CreateTable(N("a").WithType("INTEGER"))); err != nil {
		t.Error(err)
	}

	// Prepare insert statement
	insert, err := db.Prepare(N("foo").Insert("a"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log(insert)
	}

	// Insert 100 bool values
	n := 100
	v := make([]bool, n)
	for i := range v {
		v[i] = rand.Int()%2 != 0 // true or false
		if _, err := db.Exec(insert, v[i]); err != nil {
			t.Error(err)
		}
	}
	// Read rows - returns int64 values
	rows, err := db.Query(S(N("foo")))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= n; i++ {
		row := rows.Next()
		if i == n && row == nil {
			break
		} else if len(row) != 1 {
			t.Error("Unexpected row length", len(row), "in position", i)
		} else if d, ok := row[0].(int64); !ok {
			t.Errorf("Unexpected type %T in position %d", row[0], i)
		} else if (d != 0) != v[i] {
			t.Error("Unexpected value", row[0], "expected", v[i], "in position", i)
		}
	}
}

func Test_Resultset_005(t *testing.T) {
	db, err := sqlite.New(time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(N("foo").CreateTable(N("a").WithType("TIMESTAMP"))); err != nil {
		t.Error(err)
	}

	// Prepare insert statement
	insert, err := db.Prepare(N("foo").Insert("a"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log(insert)
	}

	// Insert 100 time.Time values
	n := 100
	v := make([]time.Time, n)
	for i := range v {
		v[i] = time.Now().Add(time.Second * time.Duration(rand.Int()%1e6)) // time in future, up to 1 million seconds
		if _, err := db.Exec(insert, v[i]); err != nil {
			t.Error(err)
		}
	}
	// Read rows - returns time.Time values
	rows, err := db.Query(S(N("foo")))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= n; i++ {
		row := rows.Next()
		if i == n && row == nil {
			break
		} else if len(row) != 1 {
			t.Error("Unexpected row length", len(row), "in position", i)
		} else if d, ok := row[0].(time.Time); !ok {
			t.Errorf("Unexpected type %T in position %d", row[0], i)
		} else if d.Equal(v[i]) == false {
			t.Error("Unexpected value", row[0], "expected", v[i], "in position", i)
		}
	}
}

func Test_Resultset_006(t *testing.T) {
	db, err := sqlite.New(time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(N("foo").CreateTable(N("a").WithType("DATETIME"))); err != nil {
		t.Error(err)
	}

	// Prepare insert statement
	insert, err := db.Prepare(N("foo").Insert("a"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log(insert)
	}

	// Insert 100 time.Time values
	n := 100
	v := make([]time.Time, n)
	for i := range v {
		v[i] = time.Now().Add(time.Second * time.Duration(rand.Int()%1e6)) // time in future, up to 1 million seconds
		if _, err := db.Exec(insert, v[i]); err != nil {
			t.Error(err)
		}
	}
	// Read rows - returns time.Time values
	rows, err := db.Query(S(N("foo")))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= n; i++ {
		row := rows.Next()
		if i == n && row == nil {
			break
		} else if len(row) != 1 {
			t.Error("Unexpected row length", len(row), "in position", i)
		} else if d, ok := row[0].(time.Time); !ok {
			t.Errorf("Unexpected type %T in position %d", row[0], i)
		} else if d.Equal(v[i]) == false {
			t.Error("Unexpected value", row[0], "expected", v[i], "in position", i)
		}
	}
}

func Test_Resultset_007(t *testing.T) {
	db, err := sqlite.New(time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(N("foo").CreateTable(N("a").WithType("BLOB"))); err != nil {
		t.Error(err)
	}

	// Prepare insert statement
	insert, err := db.Prepare(N("foo").Insert("a"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log(insert)
	}

	// Insert 100 []byte values
	n := 100
	v := make([][]byte, n)
	for i := range v {
		v[i] = make([]byte, rand.Int()%1e6)
		rand.Read(v[i])
		if _, err := db.Exec(insert, v[i]); err != nil {
			t.Error(err)
		}
	}
	// Read rows - returns time.Time values
	rows, err := db.Query(S(N("foo")))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= n; i++ {
		row := rows.Next()
		if i == n && row == nil {
			break
		} else if len(row) != 1 {
			t.Error("Unexpected row length", len(row), "in position", i)
		} else if d, ok := row[0].([]byte); !ok {
			t.Errorf("Unexpected type %T in position %d", row[0], i)
		} else if len(d) != len(v[i]) {
			t.Error("Unexpected length", len(d), "expected", len(v[i]), "in position", i)
		} else if hex.EncodeToString(d) != hex.EncodeToString(v[i]) {
			t.Error("Unexpected value", hex.EncodeToString(d), "expected", hex.EncodeToString(v[i]), "in position", i)
		}
	}
}
