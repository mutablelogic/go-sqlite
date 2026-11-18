package ncruces

import (
	"math"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"
	attribute "go.opentelemetry.io/otel/attribute"
)

type count uint16
type label string

func Test_argAttribute(t *testing.T) {
	n := uint32(7)
	for _, test := range []struct {
		value any
		want  attribute.KeyValue
	}{
		{"a", attribute.String("arg.v", "a")},
		{label("b"), attribute.String("arg.v", "b")},
		{true, attribute.Bool("arg.v", true)},
		{int8(-1), attribute.Int64("arg.v", -1)},
		{uint64(7), attribute.Int64("arg.v", 7)},
		{count(3), attribute.Int64("arg.v", 3)},
		{&n, attribute.Int64("arg.v", 7)},
		{uint64(math.MaxUint64), attribute.String("arg.v", "18446744073709551615")},
		{1.5, attribute.Float64("arg.v", 1.5)},
		{[]string{"a"}, attribute.String("arg.v", "[\n  \"a\"\n]")},
		{nil, attribute.String("arg.v", "null")},
	} {
		assert.Equal(t, test.want, argAttribute("v", test.value), "%#v", test.value)
	}
}
