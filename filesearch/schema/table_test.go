package schema_test

import (
	"testing"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_RankCell(t *testing.T) {
	// A rank is of three significant digits, so a small one isn't zero
	for rank, want := range map[float64]string{0.8333333: "0.833", 1.96e-6: "1.96e-06", 0.15: "0.15", 0: "0"} {
		assert.Equal(t, want, schema.Object{Rank: types.Ptr(rank)}.Cell(7), rank)
	}
	assert.Equal(t, "", schema.Object{}.Cell(7), "a list's")
}
