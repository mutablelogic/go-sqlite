package sqobj_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
	. "github.com/mutablelogic/go-sqlite/pkg/sqobj"
)

type TestVirtualStructA struct {
	A int `sqlite:"a,auto"`
}

func Test_Virtual_000(t *testing.T) {
	class := MustRegisterVirtual(N("test"), "module", TestVirtualStructA{A: 100}, "opt1", "opt2")
	t.Log(class)
}
