// Package replay provides semi-deterministic replayer.
package replay

import (
	"hash/fnv"
	"log"
	"os"
	"time"
)

// Inject calls DefaultReplayer.Inject().
func Inject(context []byte) {
	if DefaultReplayer == nil {
		// should not panic here
		log.Printf("Replayer called before initialization")
		return
	}
	DefaultReplayer.Inject(context)
}

// Replayer is an instance of Replayer.
type Replayer struct {
	// Enabled is true if this instance is enabled.
	Enabled bool
	// Debug is true if debugging mode.
	Debug bool
	// Seed is an arbitrary string for replaying execution.
	// If seed is an empty string, it disables Replayer.
	Seed string
	// Max is the max value for delays injected by Replayer.
	Max time.Duration
}

// Init initializes Replayer.
func (r *Replayer) Init() error {
	// nothing to do currently
	return nil
}

// Inject injects a random delay using context.
// The delay can be replayed with the seed value.
// Context can be nil.
func (r *Replayer) Inject(context []byte) {
	if !r.Enabled {
		return
	}
	t := r.inject(context)
	if r.Debug {
		log.Printf("t=%s (seed=%s, context=%s)",
			t, r.Seed, context)
	}
	time.Sleep(t)
}

func (r *Replayer) inject(context []byte) time.Duration {
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

func init() {
	enabled := true

	debugStr := os.Getenv("GRDEBUG")
	debug := debugStr != ""

	seed := os.Getenv("GRSEED")
	if seed == "" {
		log.Printf("GRSEED is not set. Disabling Replayer.")
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

	DefaultReplayer = &Replayer{
		Enabled: enabled,
		Debug:   debug,
		Seed:    seed,
		Max:     max,
	}
	err = DefaultReplayer.Init()
	if err != nil {
		log.Printf("Error while initializing: %s", err)
		DefaultReplayer.Enabled = false
	}
}
