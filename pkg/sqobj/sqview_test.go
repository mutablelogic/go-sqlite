package sqobj_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
	. "github.com/mutablelogic/go-sqlite/pkg/sqobj"
)

type TestView struct {
	A int
	B int
}

type TestSourceA struct {
	A int `sqlite:"a,join:a"`
	B int `sqlite:"b,join:b"`
}

type TestSourceB struct {
	A int `sqlite:"a,join:a"`
	B int `sqlite:"b,join:b"`
}

func Test_View_000(t *testing.T) {
	a := MustRegisterClass(N("TestSourceA"), TestSourceA{})
	b := MustRegisterClass(N("TestSourceB"), TestSourceB{})
	v := MustRegisterView(N("TestView"), TestView{}, false, a, b)
	t.Log(v)
}
