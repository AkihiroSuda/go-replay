// Package replay provides semi-deterministic replayer.
package replay

import (
	"hash/fnv"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Inject calls DefaultReplayer.Inject().
func Inject(context []byte) {
	if DefaultReplayer == nil {
		// should not panic here
		log.Printf("Warning: Replayer called before initialization")
		return
	}
	DefaultReplayer.Inject(context)
}

func getenv(k, defaultV string) string {
	v := os.Getenv(k)
	if v == "" {
		v = defaultV
	}
	return v
}

func probab(r *rand.Rand, p float64) bool {
	// `p.(float64) == 0` is legal in Go
	// https://github.com/golang/go/blob/go1.7.1/src/encoding/gob/encode.go#L459-L460
	if p == 0 {
		return false
	}
	return r.Float64() < p
}

type Replayer interface {
	Init() error
	Inject(context []byte)
}

// BasicReplayer is an instance of Replayer.
type BasicReplayer struct {
	random *rand.Rand
	// Debug is true if debugging mode.
	Debug bool
	// Seed is an arbitrary byte slice for replaying execution.
	// If seed is nil, it disables Replayer.
	Seed []byte
	// Max is the max value for delays injected by Replayer.
	Max time.Duration
	// ZBias is the probability of enforcing the delay to be zero.
	ZBias float64
}

// Init initializes Replayer.
func (r *BasicReplayer) Init() error {
	h := fnv.New64a()
	h.Write(r.Seed)
	ui64 := h.Sum64()
	r.random = rand.New(rand.NewSource(int64(ui64)))
	return nil
}

// Inject injects a random delay using context.
// The delay can be replayed with the seed value.
// Context can be nil.
func (r *BasicReplayer) Inject(context []byte) {
	t := r.inject(context)
	if r.Debug {
		log.Printf("t=%s (seed=%s, context=%s)",
			t, r.Seed, context)
	}
	time.Sleep(t)
}

func (r *BasicReplayer) disabled() bool {
	return r.Seed == nil || probab(r.random, r.ZBias)
}

func (r *BasicReplayer) inject(context []byte) time.Duration {
	if r.disabled() {
		return 0
	}
	h := fnv.New64a()
	h.Write(r.Seed)
	if context != nil {
		h.Write([]byte(context))
	}
	// TODO: hash runtime.Stack()
	ui64 := h.Sum64()
	t := time.Duration(ui64 % uint64(r.Max))
	return t
}

// NopReplayer is an instance of Replayer but it does nothing.
type NopReplayer struct {
}

// Init initializes Replayer.
func (r *NopReplayer) Init() error {
	// NOP
	return nil
}

// Inject injects a random delay using context.
// The delay can be replayed with the seed value.
// Context can be nil.
func (r *NopReplayer) Inject(context []byte) {
	// NOP
}

// DefaultReplayer is the default instance of Replayer.
var DefaultReplayer Replayer = &NopReplayer{}

// DefaultSeed is the default seed for DefaultReplayer.
const DefaultSeed = ""

// DefaultMax is the default max delay for DefaultReplayer.
const DefaultMax = "10ms"

// DefaultZBias is the default zbias for DefaultReplayer.
const DefaultZBias = "0.0"

func init() {
	debug := os.Getenv("GRDEBUG") != ""
	seed := os.Getenv("GRSEED")
	if seed == "" {
		if debug {
			log.Printf("GRSEED is not set. Disabling GoReplay.")
		}
		return
	}

	max, err := time.ParseDuration(getenv("GRMAX", DefaultMax))
	if err != nil {
		log.Printf("Error while parsing GRMAX: %s", err)
		return
	}

	zBias, err := strconv.ParseFloat(getenv("GRZBIAS", DefaultZBias), 64)
	if err != nil {
		log.Printf("Error while parsing GRZBIAS: %s", err)
		return
	}
	if zBias < 0.0 || zBias > 1.0 {
		log.Printf("Error: invalid GRZBIAS: %f", zBias)
		return
	}

	DefaultReplayer = &BasicReplayer{
		Debug: debug,
		Seed:  []byte(seed),
		Max:   max,
		ZBias: zBias,
	}
	if err := DefaultReplayer.Init(); err != nil {
		log.Printf("Error while initializing: %s", err)
		DefaultReplayer = &NopReplayer{}
	}
}
