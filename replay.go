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

func probab(r *rand.Rand, p float64) bool {
	return r.Float64() < p
}

// Replayer is an instance of Replayer.
type Replayer struct {
	random *rand.Rand
	// Enabled is true if this instance is enabled.
	Enabled bool
	// Debug is true if debugging mode.
	Debug bool
	// Seed is an arbitrary string for replaying execution.
	// If seed is an empty string, it disables Replayer.
	Seed string
	// Max is the max value for delays injected by Replayer.
	Max time.Duration
	// ZBias is the probability of enforcing the delay to be zero.
	ZBias float64
}

// Init initializes Replayer.
func (r *Replayer) Init() error {
	r.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	return nil
}

// Inject injects a random delay using context.
// The delay can be replayed with the seed value.
// Context can be nil.
func (r *Replayer) Inject(context []byte) {
	t := r.inject(context)
	if r.Debug {
		log.Printf("t=%s (seed=%s, context=%s)",
			t, r.Seed, context)
	}
	time.Sleep(t)
}

func (r *Replayer) inject(context []byte) time.Duration {
	if !r.Enabled {
		return 0
	}
	if r.random == nil {
		log.Printf("Warning: inject called before initialization")
		return 0
	}
	if probab(r.random, r.ZBias) {
		return 0
	}

	h := fnv.New64a()
	h.Write([]byte(r.Seed))
	if context != nil {
		h.Write([]byte(context))
	}
	// TODO: hash runtime.Stack()
	ui64 := h.Sum64()
	t := time.Duration(ui64 % uint64(r.Max))
	return t
}

// DefaultReplayer is the default instance of Replayer.
var DefaultReplayer *Replayer

// DefaultSeed is the default seed for DefaultReplayer.
const DefaultSeed = ""

// DefaultMax is the default max delay for DefaultReplayer.
const DefaultMax = "10ms"

// DefaultZBias is the default zbias for DefaultReplayer.
const DefaultZBias = "0.0"

func init() {
	enabled := true

	debugStr := os.Getenv("GRDEBUG")
	debug := debugStr != ""

	seed := os.Getenv("GRSEED")
	if seed == "" {
		if debug {
			log.Printf("GRSEED is not set. Disabling GoReplay.")
		}
		enabled = false
	}

	maxStr := os.Getenv("GRMAX")
	if maxStr == "" {
		maxStr = DefaultMax
	}
	max, err := time.ParseDuration(maxStr)
	if err != nil {
		log.Printf("Error while parsing GRMAX: %s", err)
		enabled = false
	}

	zBiasStr := os.Getenv("GRZBIAS")
	if zBiasStr == "" {
		zBiasStr = DefaultZBias
	}
	zBias, err := strconv.ParseFloat(zBiasStr, 64)
	if err != nil {
		log.Printf("Error while parsing GRZBIAS: %s", err)
		enabled = false
	}
	if zBias < 0.0 || zBias > 1.0 {
		log.Printf("Error: invalid GRZBIAS: %f", zBias)
		enabled = false
	}

	DefaultReplayer = &Replayer{
		Enabled: enabled,
		Debug:   debug,
		Seed:    seed,
		Max:     max,
		ZBias:   zBias,
	}
	err = DefaultReplayer.Init()
	if err != nil {
		log.Printf("Error while initializing: %s", err)
		DefaultReplayer.Enabled = false
	}
}
