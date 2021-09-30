package sqlite3_test

import (
	"math/rand"
	"testing"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-sqlite/pkg/sqlite3"
)

func Test_Profile_001(t *testing.T) {
	samples := NewSampleArray(5)
	t.Log(samples)
	for i := 0; i < 20; i++ {
		time.Sleep(500 * time.Millisecond)
		samples.Add(time.Duration(i) * time.Second)
		t.Log(samples)
	}
}

func Test_Profile_002(t *testing.T) {
	profiles := NewProfileArray(50, 0, 5*time.Second)
	defer profiles.Close()

	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	for i := 0; i < 200; i++ {
		time.Sleep(1000 * time.Millisecond)
		key := keys[rand.Int()%len(keys)]
		profiles.Add(key, time.Duration(rand.Int()%1000)*time.Millisecond)

		if i%20 == 0 {
			t.Log("Slow Queries:")
			for i, sample := range profiles.SlowQueries(5) {
				t.Log("   ", i, "=>", sample)
			}
		}
	}
}
