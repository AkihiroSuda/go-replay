package replay

import (
	"fmt"
	"testing"
	"time"
)

func TestInjectWithFixedSeedAndNilContext(t *testing.T) {
	r := &Replayer{
		Enabled: true,
		Debug:   true,
		Seed:    "seed",
		Max:     10 * time.Millisecond,
	}
	duration := r.inject(nil)
	for i := 0; i < 100; i++ {
		x := r.inject(nil)
		if duration != x {
			t.Fatalf("Unstable output: %s != %s", duration, x)
		}
	}
}

func TestInjectWithFixedSeedAndVariableContext(t *testing.T) {
	r := &Replayer{
		Enabled: true,
		Debug:   true,
		Seed:    "seed",
		Max:     10 * time.Millisecond,
	}
	matched := 0
	duration := r.inject(nil)
	for i := 0; i < 100; i++ {
		x := r.inject([]byte(fmt.Sprintf("i=%d", i)))
		if duration == x {
			t.Logf("match on i=%d", i)
			matched++
		}
	}
	if matched != 0 {
		t.Fatalf("matched should be zero, got %d", matched)
	}
}
