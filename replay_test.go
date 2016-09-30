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
	if err := r.Init(); err != nil {
		t.Fatal(err)
	}
	n := 100
	duration := r.inject(nil)
	for i := 0; i < n; i++ {
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
	if err := r.Init(); err != nil {
		t.Fatal(err)
	}
	zeros := 0
	n := 100
	for i := 0; i < n; i++ {
		x := r.inject([]byte(fmt.Sprintf("i=%d", i)))
		if x == 0 {
			zeros++
		}
	}
	t.Logf("n=%d, zeros=%d", n, zeros)
	if zeros != 0 {
		t.Fatalf("zeros should be 0, got %d", zeros)
	}
}

func TestInjectWithZBias(t *testing.T) {
	r := &Replayer{
		Enabled: true,
		Debug:   true,
		Seed:    "seed",
		Max:     10 * time.Millisecond,
		ZBias:   0.9,
	}
	if err := r.Init(); err != nil {
		t.Fatal(err)
	}
	zeros := 0
	n := 1000 * 1000
	for i := 0; i < n; i++ {
		x := r.inject([]byte(fmt.Sprintf("i=%d", i)))
		if x == 0 {
			zeros++
		}
	}
	t.Logf("n=%d, zBias=%f, zeros=%d", n, r.ZBias, zeros)
	if zeros == 0 || zeros == n {
		t.Fatalf("zeros should not be 0 nor %d, got %d", n, zeros)
	}
}
