package sqobj_test

import (
	"testing"
	"time"

	sqobj "github.com/djthorpe/go-sqlite/pkg/sqobj"
)

func Test_Reflect_000(t *testing.T) {
	var a struct {
		A int       `sqlite:"a,not null,primary"`
		B bool      `sqlite:"b"`
		C float32   `sqlite:"c"`
		D time.Time `sqlite:"d"`
		E []byte    `sqlite:"e"`
	}
	if q := sqobj.CreateTable("foo", &a); q == nil {
		t.Fatal("CreateTable failed")
	} else if q.Query() != "CREATE TABLE foo (a INTEGER NOT NULL,b INTEGER,c FLOAT,d TIMESTAMP,e BLOB,PRIMARY KEY (a))" {
		t.Error("Unexpected return, ", q.Query())
	}
}

func Test_Reflect_001(t *testing.T) {
	var a struct {
		A int       `sqlite:"a,index:x"`
		B bool      `sqlite:"b,index:x"`
		C float32   `sqlite:"c,unique:y"`
		D time.Time `sqlite:"d,index:z"`
		E []byte    `sqlite:"e"`
	}
	if q := sqobj.CreateIndexes("foo", &a); q == nil {
		t.Fatal("CreateIndexes failed")
	} else {
		for _, q := range q {
			t.Log(q)
		}
	}
}
