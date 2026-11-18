package ncruces

import (
	"math"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	sqlite "github.com/mutablelogic/go-sqlite"
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

// publisher ignores the changes which are published
type publisher struct{}

func (publisher) Publish(schema.Change) int { return 0 }

func Test_WriterChanges(t *testing.T) {
	// Only the writer's statements change its changes, not a reader's
	db, err := Open(t.Context(), t.TempDir()+"/test.db", sqlite.WithBroadcaster(publisher{}))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	p := db.(*poolconn).pool
	if p.writerChanges(p.writer) == nil {
		t.Error("the writer has no changes")
	}
	reader := <-p.readers
	defer func() { p.readers <- reader }()
	if p.writerChanges(reader) != nil {
		t.Error("a reader has the writer's changes")
	}
}
