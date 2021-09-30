package quote_test

import (
	"testing"

	// Import Namespace
	. "github.com/mutablelogic/go-sqlite/pkg/quote"
)

func Test_Reserved_001(t *testing.T) {
	words := ReservedWords()
	if len(words) == 0 {
		t.Error("Expected reserved words")
	}
	for _, wd := range words {
		if IsReservedWord(wd) == false {
			t.Errorf("Expected %q to be a reserved word", wd)
		}
	}
}

func Test_Reserved_002(t *testing.T) {
	types := Types()
	if len(types) == 0 {
		t.Error("Expected types")
	}
	for _, wd := range types {
		if IsType(wd) == false {
			t.Errorf("Expected %q to be a type", wd)
		}
	}
}
