package goreplay

import (
	"fmt"
	"testing"
	"time"
)

func TestInjectWithFixedSeedAndNilContext(t *testing.T) {
	gr := &GoReplay{
		Enabled: true,
		Debug:   true,
		Seed:    "seed",
		Max:     10 * time.Millisecond,
	}
	duration := gr.inject(nil)
	for i := 0; i < 100; i++ {
		x := gr.inject(nil)
		if duration != x {
			t.Fatalf("Unstable output: %s != %s", duration, x)
		}
	}
}

func TestInjectWithFixedSeedAndVariableContext(t *testing.T) {
	gr := &GoReplay{
		Enabled: true,
		Debug:   true,
		Seed:    "seed",
		Max:     10 * time.Millisecond,
	}
	matched := 0
	duration := gr.inject(nil)
	for i := 0; i < 100; i++ {
		x := gr.inject([]byte(fmt.Sprintf("i=%d", i)))
		if duration == x {
			t.Logf("match on i=%d", i)
			matched++
		}
	}
	if matched != 0 {
		t.Fatalf("matched should be zero, got %d", matched)
	}
}
