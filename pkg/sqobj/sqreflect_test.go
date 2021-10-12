package sqobj_test

import (
	"errors"
	"testing"

	// Modules
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
	. "github.com/mutablelogic/go-sqlite/pkg/sqobj"
)

type TestStruct struct {
	A int
}

func Test_Reflect_001(t *testing.T) {
	if v := ValueOf(TestStruct{}); !v.IsValid() {
		t.Error("Unexpected return from ValueOf")
	} else if v := ValueOf(&TestStruct{}); !v.IsValid() {
		t.Error("Unexpected return from ValueOf")
	} else if v := ValueOf(nil); v.IsValid() {
		t.Error("Unexpected return from ValueOf")
	} else if v := ValueOf(int(0)); v.IsValid() {
		t.Error("Unexpected return from ValueOf")
	} else if v := ValueOf([]int{}); v.IsValid() {
		t.Error("Unexpected return from ValueOf")
	}
}

func Test_Reflect_002(t *testing.T) {
	if r, err := NewReflect(TestStruct{}); err != nil {
		t.Error(err)
	} else if col := r.Column("A"); col == nil {
		t.Error("Expected column named A")
	} else if col.String() != "A INTEGER" {
		t.Error("Unexpected return:", col.String())
	}
}

type TestStructB struct {
	A int `sqlite:"a"`
	B int `sqlite:"a"`
}

func Test_Reflect_003(t *testing.T) {
	if _, err := NewReflect(TestStructB{}); errors.Is(err, ErrDuplicateEntry) {
		// Expected error
	} else {
		t.Error("Unexpected error return", err)
	}
}

type TestStructC struct {
	A int  `sqlite:"a,text,auto"`
	B bool `sqlite:"b,bool,not null"`
	C bool `sqlite:"c,primary"`
}

func Test_Reflect_004(t *testing.T) {
	r, err := NewReflect(TestStructC{A: 1, B: true, C: true})
	if err != nil {
		t.Error(err)
	}
	cola := r.Column("a")
	if decltype := cola.Type(); decltype != "TEXT" {
		t.Error("Unexpected type", decltype)
	} else if cola.String() != "a TEXT NOT NULL DEFAULT 1" {
		t.Error("Unexpected return:", cola, r)
	} else if cola.Nullable() {
		t.Error("Unexpected nullable", cola)
	} else if cola.Primary() == "" {
		t.Error("Unexpected primary", cola)
	} else {
		t.Log(cola)
	}
	colb := r.Column("b")
	if decltype := colb.Type(); decltype != "BOOL" {
		t.Error("Unexpected type", decltype)
	} else if colb.Nullable() {
		t.Error("Unexpected nullable", colb)
	} else {
		t.Log(colb)
	}
	colc := r.Column("c")
	if decltype := colc.Type(); decltype != "INTEGER" {
		t.Error("Unexpected type", decltype)
	} else if colc.Nullable() {
		t.Error("Unexpected nullable", colc)
	} else if colc.Primary() == "" {
		t.Error("Unexpected primary", colb.Primary())
	} else {
		t.Log(colc)
	}
}

type TestStructD struct {
	A int `sqlite:"a,index:a,unique:b"`
	B int `sqlite:"b,index:a,unique:b"`
}

func Test_Reflect_005(t *testing.T) {
	r, err := NewReflect(TestStructD{})
	if err != nil {
		t.Error(err)
	}
	if indexa := r.Index(N("test"), "a"); indexa == nil {
		t.Error("Expected index named a")
	} else if indexa.Query() != "CREATE INDEX test_a ON test (a,b)" {
		t.Error("Unexpected return:", indexa.Query())
	}
	if indexb := r.Index(N("test"), "b"); indexb == nil {
		t.Error("Expected index named b")
	} else if indexb.Query() != "CREATE UNIQUE INDEX test_b ON test (a,b)" {
		t.Error("Unexpected return:", indexb.Query())
	}
}

func Test_Reflect_006(t *testing.T) {
	r, err := NewReflect(TestStructD{})
	if err != nil {
		t.Error(err)
	}
	t.Logf("%q", r.Table(N("test").WithSchema("main"), true))
}

type TestStructE struct {
	A int `sqlite:"a,foreign"`
}

func Test_Reflect_007(t *testing.T) {
	r, err := NewReflect(TestStructE{})
	if err != nil {
		t.Error(err)
	}
	if err := r.WithForeignKey(N("parent"), "b"); err != nil {
		t.Error(err)
	}
	t.Log(r)
	t.Logf("%q", r.Table(N("test").WithSchema("main"), true))
}

func Test_Reflect_008(t *testing.T) {
	r, err := NewReflect(TestStructE{})
	if err != nil {
		t.Error(err)
	}
	t.Logf("%q", r.Virtual(N("test").WithSchema("main"), "module", true, "opt1", "opt2"))
}

type TestStructView struct {
	K1 int
	K2 int
}

type TestStructJoinA struct {
	K1 int `sqlite:",join:k1"`
	K2 int `sqlite:",join:k2"`
}

type TestStructJoinB struct {
	K1 int `sqlite:",join:k1"`
	K2 int `sqlite:",join:k2"`
}

func Test_Reflect_009(t *testing.T) {
	v, err := NewReflect(TestStructView{})
	if err != nil {
		t.Error(err)
	}
	// Create a new view by joining a and b together
	if view := v.View(N("test"), S(N("a")).To(N("K1"), N("K2")), true); view == nil {
		t.Error("Unexpected nil returned")
	} else {
		t.Log(view)
	}
}

func Test_Reflect_010(t *testing.T) {
	a, err := NewReflect(TestStructJoinA{})
	if err != nil {
		t.Error(err)
	}
	b, err := NewReflect(TestStructJoinB{})
	if err != nil {
		t.Error(err)
	}
	t.Log(a)
	t.Log(b)
}
